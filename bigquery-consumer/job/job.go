package job

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

type MonthlyTrend struct {
    Month        string `bigquery:"month"`
    Applications int64  `bigquery:"applications"`
    Interviews   int64  `bigquery:"interviews"`
    Offers       int64  `bigquery:"offers"`
}

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
//
// operationID (primary key)
// email (identifier for analytics and deletes)
// jobID (identifier to do job-specific analytics)
// event_time (time of event in unix seconds, required to knwow which state came first)
// applied_date (time of application in unix seconds, required to know where to place in timeline)
// status (current state of the application)
// operation (add/edit, not strictly necessary but might be useful later)
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

	// don't recalculate on userDelete
	if j.Operation == "userDelete" {
		return nil
	}

	analytics, err := j.recalculateAnalytics(ctx)
	if err != nil {
		return fmt.Errorf("failed to recalculate analytics: %w", err)
	}

	log.Printf("Analytics recalculated: %v", analytics)

	err = j.updateFirestore(ctx, analytics)
	if err != nil {
		return fmt.Errorf("failed to update Firestore: %w", err)
	}

	log.Printf("Firestore updated successfully")

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
func (j *Job) recalculateAnalytics(ctx context.Context) (map[string]interface{}, error) {
	analytics := make(map[string]interface{})

	q := j.BigQueryClient.Query(`
		-- all analytics start from latest applied date
        WITH LatestAppliedDate AS (
            SELECT
                jobID,
                MAX(applied_date) AS latest_applied_date
            FROM applications_data.applications
            WHERE email = @email
            GROUP BY jobID
        ),

		-- for each jobID, find (1) the first response aka min(event_time) after latest_applied_date
		-- and (2) the days between the two
		ResponseMetrics AS (
			SELECT
				l.jobID,
				l.latest_applied_date,
				-- use NULL to handle cases where no response yet; this ensures they are not included in the AVG
				-- and don't skew the results
				MIN(CASE 
					WHEN a.operation = 'edit' 
					AND a.status IN ('Interviewing', 'Screen', 'Offer', 'Rejected', 'Ghosted')
					AND a.event_time > l.latest_applied_date
					THEN TIMESTAMP_DIFF(a.event_time, l.latest_applied_date, DAY)
					ELSE NULL
				END) AS days_to_response
			FROM LatestAppliedDate l
			JOIN applications_data.applications a 
			ON l.jobID = a.jobID AND a.email = @email
			GROUP BY l.jobID, l.latest_applied_date
		),

		-- ANALYTIC 1 (monthly trends): # applications and interviews and offers by month
		MonthlyTrends AS (
			SELECT
				FORMAT_TIMESTAMP('%Y-%m', applied_date) AS month,
				COUNT(DISTINCT CASE WHEN OPERATION = 'add' THEN jobID ELSE NULL END) AS applications,
				COUNT(DISTINCT CASE WHEN OPERATION = 'edit' AND status = 'Interviewing' THEN jobID ELSE NULL END) AS interviews,
				COUNT(DISTINCT CASE WHEN OPERATION = 'edit' AND status = 'Offer' THEN jobID ELSE NULL END) AS offers
			FROM applications_data.applications
			WHERE email = @email
			AND applied_date >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 365 DAY)
			GROUP BY FORMAT_TIMESTAMP('%Y-%m', applied_date)
			ORDER BY month ASC
		),

		-- we can't reference the aliases in the outer query so we need to do this
        RawMetrics AS (
            SELECT
				-- ANALYTIC 2 (application velocity): # apps sent in the current 30 vs previous 30 days
				-- no need for DISTINCT because every application has at max one 'add' event
                COALESCE(
                    SUM(
						CASE
							WHEN operation = 'add'
							AND applied_date >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY) 
							THEN 1
							ELSE NULL
                    	END
					), 0 
                ) AS current_30day_count,
				COALESCE(
					SUM(
						CASE
							WHEN operation = 'add'
							AND applied_date >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 60 DAY)
							AND applied_date < TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
							THEN 1
							ELSE NULL
						END
					), 0
				) AS previous_30day_count,

				-- ANALYTIC 3 (resume effectiveness): # interviews in the current 30 vs previous 30 days
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
                ) AS previous_30day_interviews,

				-- ANALYTIC 4 (interview effectiveness): # interviews that led to offers in the current 30 vs previous 30 days
				-- agnostic of applied date, because (1) interview is not always immediate and (2) interview usually comes within 30-60 days anyway
				-- there's actually no need to actually check if previous status was 'Interviewing'
				-- because we already assume that the previous status was 'Interviewing' in the previous 30 days. plus sometimes direct offer
				COUNT(DISTINCT
					CASE WHEN a.operation = 'edit'
						AND a.status = 'Offer'
						AND a.event_time > l.latest_applied_date
						AND a.event_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
						THEN a.jobID
						ELSE NULL
					END
				) AS current_30day_offers,
				COUNT(DISTINCT
					CASE WHEN a.operation = 'edit'
						AND a.status = 'Offer'
						AND a.event_time > l.latest_applied_date
						AND a.event_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 60 DAY)
						AND a.event_time < TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
						THEN a.jobID
						ELSE NULL
					END
				) AS previous_30day_offers,

				-- ANALYTIC 5 (average response time): average time for ANY response (applied -> interivew, interview -> offer, applied -> rejected, etc)
				-- only tracks applications from current 30 or prev 30 day period
				AVG(CASE 
					WHEN l.latest_applied_date >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
					THEN rm.days_to_response
					ELSE NULL
				END) AS current_30day_avg_response_time,
				AVG(CASE
					WHEN l.latest_applied_date >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 60 DAY)
						AND l.latest_applied_date < TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
					THEN rm.days_to_response
					ELSE NULL
				END) AS previous_30day_avg_response_time
			FROM applications_data.applications a
			JOIN LatestAppliedDate l ON a.jobID = l.jobID
			LEFT JOIN ResponseMetrics rm ON a.jobID = rm.jobID
			WHERE email = @email
        )
        
        -- now we reference the aliases in an outer query
		-- COALESCE() is used to handle NULL values
        SELECT
			COALESCE(current_30day_count, 0) AS current_30day_count,
			COALESCE(previous_30day_count, 0) AS previous_30day_count,
			COALESCE(current_30day_count, 0) - COALESCE(previous_30day_count, 0) AS application_velocity_trend,
			COALESCE(current_30day_interviews, 0) AS current_30day_interviews,
			COALESCE(previous_30day_interviews, 0) AS previous_30day_interviews,
			COALESCE(current_30day_interviews, 0) - COALESCE(previous_30day_interviews, 0) AS resume_effectiveness_trend,
			COALESCE(current_30day_offers, 0) AS current_30day_offers,
			COALESCE(previous_30day_offers, 0) AS previous_30day_offers,
			COALESCE(current_30day_offers, 0) - COALESCE(previous_30day_offers, 0) AS interview_effectiveness_trend, 
			(SELECT ARRAY_AGG(STRUCT(month, applications, interviews, offers)) FROM MonthlyTrends) AS monthly_trends,
			COALESCE(current_30day_avg_response_time, 0) AS current_30day_avg_response_time,
			COALESCE(previous_30day_avg_response_time, 0) AS previous_30day_avg_response_time,
			COALESCE(current_30day_avg_response_time, 0) - COALESCE(previous_30day_avg_response_time, 0) AS response_time_trend
        FROM RawMetrics
    `)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "email", Value: j.Data["email"]},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run analytics query: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for job: %w", err)
	}

	if err := status.Err(); err != nil {
		return nil, fmt.Errorf("job completed with error: %w", err)
	}

	it, err := job.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read analytics data: %w", err)
	}

	// rows may be null so we use an int pointer
	var row struct {
		Current30DayCount       int `bigquery:"current_30day_count"`
		Previous30DayCount      int `bigquery:"previous_30day_count"`
		ApplicationVelocityTrend int `bigquery:"application_velocity_trend"`
		Current30DayInterviews  int `bigquery:"current_30day_interviews"`
		Previous30DayInterviews int `bigquery:"previous_30day_interviews"`
		Current30DayOffers 	int `bigquery:"current_30day_offers"`
		Previous30DayOffers 	int `bigquery:"previous_30day_offers"`
		InterviewEffectivenessTrend int `bigquery:"interview_effectiveness_trend"`
		ResumeEffectivenessTrend int `bigquery:"resume_effectiveness_trend"`
		Current30DayAvgResponseTime float64 `bigquery:"current_30day_avg_response_time"`
		Previous30DayAvgResponseTime float64 `bigquery:"previous_30day_avg_response_time"`
		ResponseTimeTrend float64 `bigquery:"response_time_trend"`
		MonthlyTrends           []MonthlyTrend `bigquery:"monthly_trends"`
	}

	if err := it.Next(&row); err != nil {
		if err == iterator.Done {
			// no analyitcs found
			return analytics, nil
		}
		return nil, fmt.Errorf("failed to read row: %w", err)
	}

	analytics["application_velocity"] = row.Current30DayCount
	analytics["application_velocity_trend"] = row.ApplicationVelocityTrend
	analytics["resume_effectiveness"] = row.Current30DayInterviews
	analytics["resume_effectiveness_trend"] = row.ResumeEffectivenessTrend
	analytics["monthly_trends"] = row.MonthlyTrends
	analytics["avg_response_time"] = row.Current30DayAvgResponseTime
	analytics["avg_response_time_trend"] = row.ResponseTimeTrend
	analytics["interview_effectiveness"] = row.Current30DayOffers
	analytics["interview_effectiveness_trend"] = row.InterviewEffectivenessTrend
	analytics["last_updated"] = time.Now().Unix()

	return analytics, nil
}

// update the Firestore document with the new analytics
func (j *Job) updateFirestore(ctx context.Context, analytics map[string]interface{}) error {
	doc := j.FirestoreClient.Collection("users").Doc(j.Data["email"].(string))

	_, err := doc.Set(ctx, analytics, firestore.MergeAll)
	if err != nil {
		return fmt.Errorf("failed to update Firestore document: %w", err)
	}

	return nil
}