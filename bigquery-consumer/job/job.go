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
	case "revert": 
		err = j.revert(ctx)
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


// reverts only the most recent operation for a given job ID. Algolia and Firestore do not know the UUIDs made in BigQuery
// so we have to rely on event_time. However, Go enforces (1) no duplicate status updates and (2) no duplicate reverts
// so this is actually safe to do
func (j *Job) revert(ctx context.Context) error {
	q := j.BigQueryClient.Query(`
		UPDATE applications_data.applications
		SET operation = 'revert'
		WHERE email = @email
		AND jobID = @jobID
		AND event_time = (
			SELECT MAX(event_time) 
			FROM applications_data.applications
			WHERE email = @email
			AND jobID = @jobID
		)
	`)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "email", Value: j.Data["email"]},
		{Name: "jobID", Value: j.Data["objectID"]},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to revert record: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for job: %w", err)
	}

	if err := status.Err(); err != nil {
		return fmt.Errorf("job completed with error: %w", err)
	}

	log.Printf("Job [%v] reverted successfully for email [%v]", j.Data["objectID"], j.Data["email"])

	return nil
}

// each key in the map is the name of the analytic (identical to Firestore field name)
// this is so that we can easily add more analytics in the future if we want
// let's batch run all queries instead of running them one by one
// by batch run I mean a huge query that does every calculation lol
func (j *Job) recalculateAnalytics(ctx context.Context) (map[string]interface{}, error) {
	analytics := make(map[string]interface{})

	q := j.BigQueryClient.Query(`
		-- extract all relevant data in one pass of applications table
		WITH UserApplications AS (
			SELECT 
				jobID,
				email,
				event_time,
				applied_date,
				status,
				operation,
				-- pre-calculate conditions we'll use multiple times
				(operation = 'add') AS is_application,
				(operation = 'edit' AND status IN ('Interviewing', 'Screen')) AS is_interview,
				(operation = 'edit' AND status = 'Offer') AS is_offer,
				(operation = 'edit' AND status IN ('Interviewing', 'Screen', 'Offer', 'Rejected', 'Ghosted')) AS is_response,
				-- time periods
				(applied_date >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)) AS in_current_period,
				(applied_date >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 60 DAY) 
				AND applied_date < TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)) AS in_previous_period,
				FORMAT_TIMESTAMP('%Y-%m', applied_date) AS month
			FROM applications_data.applications
			WHERE email = @email
			AND operation != 'revert'
		),

		-- latest applied date must be separate from JobMetrics (window functions cannot be nested within aggregate functions)
		LatestAppliedDates AS (
			SELECT
				jobID,
				MAX(applied_date) AS latest_applied_date
			FROM UserApplications
			GROUP BY jobID
		),

		-- calculate avg time to first response 
		ResponseMetrics AS (
			SELECT
				l.jobID,
				l.latest_applied_date,
				MIN(CASE 
					WHEN ua.is_response AND ua.event_time > l.latest_applied_date
					THEN TIMESTAMP_DIFF(ua.event_time, l.latest_applied_date, DAY)
					ELSE NULL
				END) AS days_to_response
			FROM LatestAppliedDates l
			JOIN UserApplications ua ON l.jobID = ua.jobID
			GROUP BY l.jobID, l.latest_applied_date
		),

		-- get monthly trends in # apps sent, # interview, # offer
		MonthlyTrends AS (
			SELECT
				month,
				SUM(CASE WHEN is_application THEN 1 ELSE 0 END) AS applications,
				COUNT(DISTINCT CASE WHEN is_interview THEN jobID END) AS interviews,
				COUNT(DISTINCT CASE WHEN is_offer THEN jobID END) AS offers
			FROM UserApplications
			WHERE applied_date >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 365 DAY)
			GROUP BY month
			ORDER BY month
		)

		-- final select to aggregate all metrics into one row
		-- Q: some subqueries are made here; but why not JOIN them?
		-- A: subqueries are just one operation -- none of the data from UserApplications is needed
		--    to process ResponseMetrics or MonthlyTrends. Furthermore, columnar storage + BigQuery's optimizer
		--    will likely recognize that subqueries can be run independently and processed in parallel
		SELECT
			-- application velocity metrics
			COALESCE(SUM(CASE WHEN ua.in_current_period AND ua.is_application THEN 1 ELSE 0 END), 0) AS current_30day_count,
			COALESCE(SUM(CASE WHEN ua.in_previous_period AND ua.is_application THEN 1 ELSE 0 END), 0) AS previous_30day_count,
			COALESCE(SUM(CASE WHEN ua.in_current_period AND ua.is_application THEN 1 ELSE 0 END), 0) - 
				COALESCE(SUM(CASE WHEN ua.in_previous_period AND ua.is_application THEN 1 ELSE 0 END), 0) AS application_velocity_trend,
			
			-- resume effectiveness metrics
			COUNT(DISTINCT CASE WHEN ua.event_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY) 
							AND ua.is_interview THEN ua.jobID ELSE NULL END) AS current_30day_interviews,
			COUNT(DISTINCT CASE WHEN ua.event_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 60 DAY)
							AND ua.event_time < TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
							AND ua.is_interview THEN ua.jobID ELSE NULL END) AS previous_30day_interviews,
			COUNT(DISTINCT CASE WHEN ua.event_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY) 
							AND ua.is_interview THEN ua.jobID ELSE NULL END) -
			COUNT(DISTINCT CASE WHEN ua.event_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 60 DAY)
							AND ua.event_time < TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
							AND ua.is_interview THEN ua.jobID ELSE NULL END) AS resume_effectiveness_trend,
			
			-- interview effectiveness metrics
			COUNT(DISTINCT CASE WHEN ua.event_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY) 
							AND ua.is_offer THEN ua.jobID ELSE NULL END) AS current_30day_offers,
			COUNT(DISTINCT CASE WHEN ua.event_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 60 DAY)
							AND ua.event_time < TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
							AND ua.is_offer THEN ua.jobID ELSE NULL END) AS previous_30day_offers,
			COUNT(DISTINCT CASE WHEN ua.event_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY) 
							AND ua.is_offer THEN ua.jobID ELSE NULL END) -
			COUNT(DISTINCT CASE WHEN ua.event_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 60 DAY)
							AND ua.event_time < TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
							AND ua.is_offer THEN ua.jobID ELSE NULL END) AS interview_effectiveness_trend,
			
			-- response time metrics (gets as a subquery, avoids unnecessary JOIN since does not use UserApplications)
			-- scalar subqueries can only have one column so separate into multiple
			(SELECT
				AVG(CASE
						WHEN latest_applied_date >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
						THEN days_to_response
						ELSE NULL
					END)
			FROM ResponseMetrics) AS current_30day_avg_response_time,
			(SELECT
				AVG(CASE
						WHEN latest_applied_date >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 60 DAY)
						AND latest_applied_date < TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
						THEN days_to_response
						ELSE NULL
					END)
			FROM ResponseMetrics) AS previous_30day_avg_response_time,
			(SELECT
				AVG(CASE
						WHEN latest_applied_date >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
						THEN days_to_response
						ELSE NULL
					END) - 
				AVG(CASE WHEN latest_applied_date >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 60 DAY)
						AND latest_applied_date < TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
						THEN days_to_response
						ELSE NULL 
					END)
			FROM ResponseMetrics) AS response_time_trend,
			
			-- monthly trends (gets as a subquery, avoids unnecessary JOIN since does not use UserApplications)
			(SELECT
				ARRAY_AGG(
					STRUCT(month, applications, interviews, offers)
				)
			FROM MonthlyTrends) AS monthly_trends

		FROM UserApplications ua
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
		Current30DayAvgResponseTime bigquery.NullFloat64 `bigquery:"current_30day_avg_response_time"`
		Previous30DayAvgResponseTime bigquery.NullFloat64 `bigquery:"previous_30day_avg_response_time"`
		ResponseTimeTrend bigquery.NullFloat64 `bigquery:"response_time_trend"`
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
	if row.Current30DayAvgResponseTime.Valid {
		analytics["avg_response_time"] = row.Current30DayAvgResponseTime.Float64
	} else {
		analytics["avg_response_time"] = nil 
	}
	
	if row.ResponseTimeTrend.Valid {
		analytics["avg_response_time_trend"] = row.ResponseTimeTrend.Float64
	} else {
		analytics["avg_response_time_trend"] = nil
	}
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