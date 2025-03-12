package job 

import (
	"encoding/json"
	"fmt"
	"context"
	"log"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/firestore"
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

	// recalculates user's analytics (with queries) and updates their analytics fields in Firestore
	// 1. (total volume) trend in applications submitted over the past 30 days vs the previous 30 days
	// 2. (resume effectiveness)
	// 		a. trend in applications converting to interview stage over the past 30 days vs the previous 30 days
	//			- this is ANY application where event_time is in the past (or previous) 30 days and status is interviewing
	//			- so this isn't just limited to applications that were submitted within the past 30 days
	//		b. trend in applications converting to rejection stage over the past 30 days vs the previous 30 days
	//			- same as above, but for rejection stage
	// 3. (interview effectiveness) trend in applications converting from interview to offer stage over the past 30 days vs the previous 30 days
	//		- same as above, but for offer stage
	// 4. (response time) avg time to first response over the past 60 days
	//		- this one is a different kind of analytic, not a trend but an average. basically looks at all applications
	//	      sent in the past 60 days and calculates time between application date and first status change
	// 5. (status progression velocity) avg time spent in one stage before moving to the next over the past 60 days
	//		- this one is a different kind of analytic, not a trend but an average
	// 6. (improvement over time) trend in number of rejection vs interview/offer over the past 30 days vs the previous 30 days
	//		- this is a ratio of rejections to interviews/offers and the trend in that ratio
	//		- this can be inferred from 2.a and 2.b but it's a useful metric to have
	// 7. (best month all time) identify month with the most applications converting to interview/offer
	// NOTE: for efficiency, we should batch these queries!!! and also we ``might`` be able to statically analyze
	//		 which queries don't need to be re-ran based on data 
	//		 - e.g. if event timestamp is beyond 60 days old, no need to re-run 1-6
	//		 - if event contains status 'interviewing' then 2.b doesn't need to be re-ran and so on
	// err = j.recalculateAnalyticsAndUpdateFirestore()
	// if err != nil {
	// 	return fmt.Errorf("failed to recalculate analytics: %w", err)
	// }

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

// delete anything matching this user and the job ID (chance that job ID is not unique)
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