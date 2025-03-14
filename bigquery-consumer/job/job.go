package job 

import (
	"encoding/json"
	"fmt"
	"context"
	"log"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"github.com/google/uuid"
)

type Job struct {
    ID              int32
    Data            map[string]interface{}
    RawData         []byte
    Operation       string
    BigQueryClient  *bigquery.Client
    FirestoreClient *firestore.Client
}

// all this really does is unmarshal the raw data and figure out the operation 
func NewJob(data []byte, id int32, bqClient *bigquery.Client, fsClient *firestore.Client) (*Job, error) {
    var parsedData map[string]interface{}
    err := json.Unmarshal(data, &parsedData)
    if err != nil {
        return nil, fmt.Errorf("failed to parse job data: %w", err)
    }
    
    operation, ok := parsedData["operation"].(string)
    if !ok {
        return nil, fmt.Errorf("missing or invalid operation field")
    }
    
    return &Job{
        ID:              id,
        RawData:         data,
        Data:            parsedData,
        Operation:       operation,
        BigQueryClient:  bqClient,
        FirestoreClient: fsClient,
    }, nil
}

// Process handles the job based on its operation type
// dataset: applications_data
// table: applications
// schema:
// 	operationID (primary key)
//  email (identifier for analytics and deletes)
//  jobID (identifier to do job-specific analytics)
//	event_time (time of event in unix ms, required to knwow which state came first)
//	applied_date (time of application in unix ms, required to know where to place in timeline)
//	status (current state of the application)
//	operation (add/edit, not strictly necessary but might be useful later)
func (j *Job) Process(ctx context.Context) error {
    log.Printf("[*] BigQuery [*]")
    log.Printf("-------")
    log.Printf("Processing Job With ID [%d] with content: [%s]", j.ID, j.Data)

	var err error
    
    switch j.Operation {
    case "add", "edit":
        err = j.appendJob(ctx)
    case "delete":
        err = j.deleteJob(ctx)
    case "userDelete":
        err = j.deleteUser(ctx)
    default:
        err = fmt.Errorf("unknown operation: %s", j.Operation)
    }

	if err != nil {
		return fmt.Errorf("failed to process job: %w", err)
	}

	analytics, err := j.recalculateAnalytics()
	if err != nil {
		return fmt.Errorf("failed to recalculate analytics: %w", err)
	}

	log.Printf("Analytics recalculated: %v", analytics)

	return nil
}

// appends a job to the applications table
func (j *Job) appendJob(ctx context.Context) error {
	q := j.BigQueryClient.Query(`
		INSERT INTO applications_data.applications 
  			(operationID, email, jobID, event_time, applied_date, status, operation)
		VALUES 
  			(@operationID,
			@email,
			@jobID,
			TIMESTAMP_SECONDS(@event_time),
			TIMESTAMP_SECONDS(@applied_date),
			@status,
			@operation
		)
	`)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "operationID", Value: uuid.New().String()},
		{Name: "email", Value: j.Data["email"]},
		{Name: "jobID", Value: j.Data["objectID"]},
		// json assumes all numbers are floats, so we need to cast to int64 (our schema requires it)
		{Name: "event_time", Value: int64(j.Data["timestamp"].(float64))},
		{Name: "applied_date", Value: int64(j.Data["appliedDate"].(float64))},
		{Name: "status", Value: j.Data["status"]},
		{Name: "operation", Value: j.Operation},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to insert record: %w", err)
	}

	status, err := job.Wait(ctx)
    if err != nil {
        return fmt.Errorf("failed to wait for job: %w", err)
    }
    
    if err := status.Err(); err != nil {
        return fmt.Errorf("job completed with error: %w", err)
	}

	log.Printf("Job [%v] inserted successfully with UUID [%v]", j.Data["objectID"], q.Parameters[0].Value)

	return nil
}

// delete anything matching this user and the job ID (chance that job ID is not unique so we also need email)
func (j *Job) deleteJob(ctx context.Context) error {
	q := j.BigQueryClient.Query(`
		DELETE FROM applications_data.applications
		WHERE email = @email
		AND jobID = @jobID
	`)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "email", Value: j.Data["email"]},
		{Name: "jobID", Value: j.Data["objectID"]},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for job: %w", err)
	}

	if err := status.Err(); err != nil {
		return fmt.Errorf("job completed with error: %w", err)
	}

	log.Printf("Job [%v] deleted successfully for email [%v]", j.Data["objectID"], j.Data["email"])

	return nil
}

// delete all records matching the user ID
func (j *Job) deleteUser(ctx context.Context) error {
	q := j.BigQueryClient.Query(`
		DELETE FROM applications_data.applications
		WHERE email = @email
	`)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "email", Value: j.Data["email"]},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for job: %w", err)
	}

	if err := status.Err(); err != nil {
		return fmt.Errorf("job completed with error: %w", err)
	}

	log.Printf("All jobs deleted successfully for email [%v]", j.Data["email"])
	
	return nil
}

// each key in the map is the name of the analytic (identical to Firestore field name)
// this is so that we can easily add more analytics in the future if we want
// let's batch run all queries instead of running them one by one
// by batch run I mean a huge query that does every calculation lol
func (j *Job) recalculateAnalytics() (map[string]interface{}, error) {
	analytics := make(map[string]interface{})

    q := j.BigQueryClient.Query(`
        WITH LatestAppliedDate AS (
            SELECT
                jobID,
                MAX(applied_date) AS latest_applied_date
            FROM applications_data.applications
            WHERE email = @email
            GROUP BY jobID
        ),

		-- we can't reference the aliases in the outer query so we need to do this
        RawMetrics AS (
            SELECT
				-- ANALYTIC 1 (application velocity): # apps sent in the current 30 vs previous 30 days
                SUM(
                    CASE
                        WHEN operation = 'add'
                        AND applied_date >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY) 
                        THEN 1
                        ELSE 0 
                    END
                ) AS current_30day_count,
                SUM(
                    CASE
                        WHEN operation = 'add'
                        AND applied_date >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 60 DAY)
                        AND applied_date < TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
                        THEN 1
                        ELSE 0
                    END
                ) AS previous_30day_count,

				-- ANALYTIC 2 (resume effectiveness): # interviews in the current 30 vs previous 30 days
				-- agnostic of applied date, because (1) interview is not always immediate and (2) interview usually comes within 30-60 days anyway
                COUNT(DISTINCT 
                    CASE WHEN a.operation = 'edit'
                        AND (a.status = 'Interviewing' OR a.status = 'Screen')
                        AND a.event_time > l.latest_applied_date
                        AND a.event_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
                        THEN a.jobID
                        ELSE NULL
                    END
                ) AS current_30day_interviews,
                COUNT(DISTINCT
                    CASE WHEN a.operation = 'edit'
                        AND (a.status = 'Interviewing' OR a.status = 'Screen')
                        AND a.event_time > l.latest_applied_date
                        AND a.event_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 60 DAY)
                        AND a.event_time < TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
                        THEN a.jobID
                        ELSE NULL
                    END
                ) AS previous_30day_interviews
            FROM applications_data.applications a
            JOIN LatestAppliedDate l ON a.jobID = l.jobID
            WHERE email = @email
            AND applied_date >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 60 DAY)
        )
        
        -- now we reference the aliases in an outer query
		-- COALESCE() is used to handle NULL values
        SELECT
			COALESCE(current_30day_count, 0) AS current_30day_count,
			COALESCE(previous_30day_count, 0) AS previous_30day_count,
			COALESCE(current_30day_count, 0) - COALESCE(previous_30day_count, 0) AS application_velocity,
			COALESCE(current_30day_interviews, 0) AS current_30day_interviews,
			COALESCE(previous_30day_interviews, 0) AS previous_30day_interviews,
			COALESCE(current_30day_interviews, 0) - COALESCE(previous_30day_interviews, 0) AS resume_effectiveness
        FROM RawMetrics
    `)

	// TODO: any all time analytics should be done in a separate query because....
	// 		the AND applied_Date >= ... is used in prev query because
	//		the WHERE statement is executed before the SELECT clause which is a little optimization to never
	// 		look at records that are older than 60 days
	q.Parameters = []bigquery.QueryParameter{
		{Name: "email", Value: j.Data["email"]},
	}

	it, err := q.Read(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to read analytics data: %w", err)
	}

	// rows may be null so we use an int pointer
	var row struct {
		Current30DayCount int `bigquery:"current_30day_count"`
		Previous30DayCount int `bigquery:"previous_30day_count"`
		ApplicationVelocity int `bigquery:"application_velocity"`
		Current30DayInterviews int `bigquery:"current_30day_interviews"`
		Previous30DayInterviews int `bigquery:"previous_30day_interviews"`
		ResumeEffectiveness int `bigquery:"resume_effectiveness"`
	}

	if err := it.Next(&row); err != nil {
		if err == iterator.Done {
			// no analyitcs found
			return analytics, nil
		}
		return nil, fmt.Errorf("failed to read row: %w", err)
	}

	analytics["application_velocity"] = row.ApplicationVelocity
	analytics["resume_effectiveness"] = row.ResumeEffectiveness

	return analytics, nil
}