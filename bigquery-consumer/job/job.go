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
func (j *Job) Process() error {
    log.Printf("[*] BigQuery [*]")
    log.Printf("-------")
    log.Printf("Processing Job With ID [%d] with content: [%s]", j.ID, j.Data)

	var err error
    
    switch j.Operation {
    case "add", "edit":
        err = j.appendJob()
    case "delete":
        err = j.deleteJob()
    case "userDelete":
        err = j.deleteUser()
    default:
        err = fmt.Errorf("unknown operation: %s", j.Operation)
    }

	if err != nil {
		return fmt.Errorf("failed to process job: %w", err)
	}

	// err = j.recalculateAnalytics()
	// if err != nil {
	// 	return fmt.Errorf("failed to recalculate analytics: %w", err)
	// }

	return nil
}

// appends a job to the applications table
func (j *Job) appendJob() error {
	q := j.BigQueryClient.Query(`
		INSERT INTO applications_data.applications 
  			(operationID, email, jobID, event_time, applied_date, status, operation)
		VALUES 
  			(@operationID, @email, @jobID, @event_time, @applied_date, @status, @operation)
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

	job, err := q.Run(context.Background())
	if err != nil {
		return fmt.Errorf("failed to insert record: %w", err)
	}

	status, err := job.Wait(context.Background())
    if err != nil {
        return fmt.Errorf("failed to wait for job: %w", err)
    }
    
    if err := status.Err(); err != nil {
        return fmt.Errorf("job completed with error: %w", err)
    }

	log.Println("Job inserted successfully")

	return nil
}

func (j *Job) deleteJob() error {
	// delete all records matching the job ID
	return nil
}

func (j *Job) deleteUser() error {
	// delete all records matching the user ID
	// ...
	return nil
}