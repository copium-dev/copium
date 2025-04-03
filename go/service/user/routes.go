package user

// this file contains the HTTP handlers for the user service
// it contains the following handlers:
// (R) - Dashboard: queries Postgres w/ search query for applications
// (R) - Profile: queries Postgres for user profile data and their analytics
// (C) - AddApplication: adds an application to Postgres, updates applied count and application velocity
// (D) - DeleteApplication: deletes an application from Postgres, fully recalculates analytics
// (U) - EditStatus: edits status of application, incr/decr old and new status counts and selectively updates analytics
// (U) - EditApplication: simply edit application metadata, no analytics updates
// (D) - DeleteUser: just cascade delete the user, analytics updates will follow
// this file used to have crazy pubsub logic with compensating transactions but after
// a full postgres migration, that's no longer necessary!

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"database/sql"
	"math"

	"github.com/copium-dev/copium/go/service/auth"
	"github.com/copium-dev/copium/go/service/user/userutils"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/jackc/pgx/v5"
)

// just some stupid way to make an enum and implement the json unmarshaler interface
type ApplicationStatus = userutils.ApplicationStatus

type AddApplicationRequest struct {
	Role        string `json:"role"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	AppliedDate int64  `json:"appliedDate"`
	Status      ApplicationStatus `json:"status"`
	Link        string `json:"link"`
}

type Application struct {
	ID          string `json:"id"`
	Role        string `json:"role"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	AppliedDate int64  `json:"appliedDate"`
	Status      ApplicationStatus `json:"status"`
	Link        string `json:"link"`
}

type Operation struct {
	OperationID string `json:"operationID"`
	Operation   string `json:"operation"`
	Status      string `json:"status"`
	EventTime   time.Time  `json:"event_time"`
}

type DashboardResponse struct {
	Applications []SearchResponse `json:"applications"`
	TotalPages   int64               `json:"totalPages"`
	CurrentPage  int               `json:"currentPage"`
}

type SearchResponse struct {
	ID          string `json:"objectID"`
	Role        string `json:"role"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	AppliedDate int64  `json:"appliedDate"`
	Status      ApplicationStatus `json:"status"`
	Link        string `json:"link"`
}

type DeleteApplicationRequest struct {
	ID          string `json:"id"`
	Role        string `json:"role"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	AppliedDate int64  `json:"appliedDate"`
	Status      ApplicationStatus `json:"status"`
	Link        string `json:"link"`
}


type EditApplicationStatusRequest struct {
	ID          string `json:"id"`
	Status      ApplicationStatus `json:"status"`
}

// edit application does not include status because status is edited separately
// as of postgres migration, no need to send old values to perform compensating
// transactions because this is just a DB update... yay
type EditApplicationRequest struct {
	ID             string `json:"id"`
	Role           string `json:"role"`
	Company        string `json:"company"`
	Location       string `json:"location"`
	Link           string `json:"link"`
}

type RevertApplicationStatusRequest struct {
	OperationID string `json:"operationID"`
	ID       string `json:"id"`
}

type ApplicationTimelineRequest struct {
	ID string `json:"id"`
}

type Handler struct {
	pgClient *pgxpool.Pool
	redisClient *redis.Client
}

func NewHandler(pgClient *pgxpool.Pool, redisClient *redis.Client) *Handler {
	return &Handler{
		pgClient: pgClient,
		redisClient: redisClient,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/user/dashboard", h.Dashboard).Methods("GET").Name("dashboard")
	router.HandleFunc("/user/profile", h.Profile).Methods("GET").Name("profile")
	router.HandleFunc("/user/addApplication", h.AddApplication).Methods("POST").Name("addApplication")
	router.HandleFunc("/user/deleteApplication", h.DeleteApplication).Methods("POST").Name("deleteApplication")
	router.HandleFunc("/user/editStatus", h.EditStatus).Methods("POST").Name("editStatus")
	router.HandleFunc("/user/editApplication", h.EditApplication).Methods("POST").Name("editApplication")
	router.HandleFunc("/user/deleteUser", h.DeleteUser).Methods("POST").Name("deleteUser")
	router.HandleFunc("/user/revertStatus", h.RevertStatus).Methods("POST").Name("revertStatus")
	router.HandleFunc("/user/getApplicationTimeline", h.GetApplicationTimeline).Methods("POST").Name("getApplicationTimeline")
}

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	log.Println("[*] Profile [*]")
	log.Println("-----------------")

	email, err := auth.IsAuthenticated(r)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Println("User authenticated")

	res, err := h.extractAnalytics(r.Context(), email)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Error extracting analytics", http.StatusInternalServerError)
		return
	}

	log.Println("Profile data extracted")
	log.Println("-----------------")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// does FTS search on applications table (if query provided)
// plus other filters
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	log.Println("[*] Dashboard [*]")
	log.Println("-----------------")

	email, err := auth.IsAuthenticated(r)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Println("User authenticated")

	// extract params 
	params := r.URL.Query()
	queryText := params.Get("q")

	log.Println("Query text:", queryText)

	// extract page number from query params
	pageStr := r.URL.Query().Get("page")
	page := 0
	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		// since frontend always sends 1-indexed page number, subtract by 1
		if err == nil && p > 0 {
			page = p - 1
		}
	}
	log.Println("Page requested:", page)
	
	// get latest cache version for this user. we use cache versions to avoid having
	// to do expensive cache deletes; we just let the old one expire
	versionKey := fmt.Sprintf("user:%s:cache_version", email)
    version, err := h.redisClient.Get(r.Context(), versionKey).Int64()
    if err == redis.Nil {
		// no version exists, create new one 
        version = 1
        h.redisClient.Set(r.Context(), versionKey, version, 0) // no expiration
    }

	// get previous page's time boundary
	boundaryKey := fmt.Sprintf("user:%s:page:%d:%d", email, page-1, version)
    val, err := h.redisClient.Get(r.Context(), boundaryKey).Result()
    
    var prevPageBoundary *time.Time
	if err == nil {
		fmt.Println("Previous page boundary cached:", val)
		// storing as milliseconds in redis, so convert to time.UnixMilli()
		// we can interchangeably use timestamptz (Postgres) and time.Time (Go)
		millis, err := strconv.ParseInt(val, 10, 64)
		if err == nil {
			t := time.UnixMilli(millis)
			prevPageBoundary = &t
		}
	}

	hitsPerPage := r.URL.Query().Get("hits")
	hitsPerPageInt := 10

	if hitsPerPage != "" {
		parsed, err := strconv.Atoi(hitsPerPage)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			http.Error(w, "Error parsing hitsPerPage", http.StatusBadRequest)
			return
		}
		hitsPerPageInt = parsed
	}

	// 1. email
	// 2. query text (full text on title, company, locations)
	// 3. limit
	// 4. offset (aka page requested)
	// 5. prevPageBoundary (optional, for pagination)
	applications, totalHits, err := h.filterQueryText(email, queryText, r.Context(), hitsPerPageInt, page, prevPageBoundary)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Error filtering applications", http.StatusInternalServerError)
		return
	}

	// set new boundary (if any) for the current page, also make sure we're on the right cache version
	if len(applications) > 0 {
		// recall that we need millisecond precision for keyset pagination
		// thankfully we already store millisecond precision in database
        h.redisClient.Set(r.Context(), 
            fmt.Sprintf("user:%s:page:%d:%d", email, page, version),
			fmt.Sprintf("%d", applications[len(applications)-1].AppliedDate),
            time.Hour) // 1 hour TTL
    }
	
	log.Println("Applications extracted:", applications)
	log.Println("-----------------")

	log.Println("total hits:", totalHits)

	// create response object pagination info
	responseObject := DashboardResponse{
		Applications: applications,
		// so ugly sorry but yeah gotta do this
		TotalPages: int64(math.Ceil(float64(totalHits) / float64(hitsPerPageInt))),
		CurrentPage:  page,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseObject)
}

// consistency: if publish fails, delete the application from Firestore
func (h *Handler) AddApplication(w http.ResponseWriter, r *http.Request) {
	log.Println("[*] AddApplication [*]")
	log.Println("-----------------")

	email, err := auth.IsAuthenticated(r)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Println("User authenticated")

	var addApplicationRequest AddApplicationRequest
	err = json.NewDecoder(r.Body).Decode(&addApplicationRequest)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	fmt.Println("AddApplicationRequest:", addApplicationRequest)

	// extract json from request body
	var newAppID sql.NullString
	// note: postgres to_timestamp expects seconds so convert here, but we use floating point
	// to preserve millisecond precision by storing it in the decimal
	err = h.pgClient.QueryRow(r.Context(), `
		SELECT service.add_application($1, to_timestamp(($2::BIGINT)/1000.0)::timestamptz, $3, $4, $5, $6, $7, $8)
	`, 
		email,                                      // p_email
		addApplicationRequest.AppliedDate,         	// p_applied_date (as Unix timestamp, convert to timestamp)
		"add",
		addApplicationRequest.Status,               // p_app_status
		addApplicationRequest.Company,              // p_company
		addApplicationRequest.Role,                 // p_title (role)
		addApplicationRequest.Link,                 // p_link
		addApplicationRequest.Location,             // p_locations
	).Scan(&newAppID)
	if err != nil {
		fmt.Printf("Error calling add_application: %v\n", err)
		http.Error(w, "Error adding application", http.StatusInternalServerError)
		return
	}

	objectID := newAppID.String
	if !newAppID.Valid {
		fmt.Println("Error: add_application returned null ID")
		http.Error(w, "Error adding application", http.StatusInternalServerError)
		return
	}

	// 'invalidate' aka increment cache so user has a new cache version
	h.invalidateUserCache(r.Context(), email)

	log.Println("Application added")
	log.Println("-----------------")

	// return id for optimsitic UI
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"objectID": objectID,
	})
}

// consistency: save current application status; if publish fails, revert status
func (h *Handler) DeleteApplication(w http.ResponseWriter, r *http.Request) {
	log.Println("[*] DeleteApplication [*]")
	log.Println("-----------------")

	email, err := auth.IsAuthenticated(r)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Println("User authenticated")

	// extract json from request body
	var deleteApplicationRequest DeleteApplicationRequest
	err = json.NewDecoder(r.Body).Decode(&deleteApplicationRequest)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	applicationID := deleteApplicationRequest.ID

	var success bool
	err = h.pgClient.QueryRow(r.Context(), 
		`SELECT service.delete_application($1, $2)`, email, applicationID,
	).Scan(&success)
	if err != nil {
		fmt.Printf("Error calling delete_full_recalculate_analytics: %v\n", err)
		http.Error(w, "Error deleting application", http.StatusInternalServerError)
		return
	}

	if !success {
		fmt.Println("Error: delete_full_recalculate_analytics failed")
		http.Error(w, "Error deleting application", http.StatusInternalServerError)
		return
	}

	// 'invalidate' aka increment cache so user has a new cache version
	h.invalidateUserCache(r.Context(), email)

	log.Println("Application deleted, logs recalculated")
	log.Println("-----------------")

	w.WriteHeader(http.StatusOK)
}

// NOTE: eager loading is not necessary here because frontend already assumes success
// if any error, a refresh would show the correct (reverted) state
// does not invalidate cache because applied date does not change
func (h *Handler) EditStatus(w http.ResponseWriter, r *http.Request) {
	log.Println("[*] EditStatus [*]")
	log.Println("-----------------")

	email, err := auth.IsAuthenticated(r)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Println("User authenticated")

	// extract json from request body
	var EditApplicationStatusRequest EditApplicationStatusRequest
	err = json.NewDecoder(r.Body).Decode(&EditApplicationStatusRequest)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	fmt.Println("EditApplicationStatusRequest:", EditApplicationStatusRequest)

	applicationID := EditApplicationStatusRequest.ID
	newStatus := EditApplicationStatusRequest.Status

	var success bool
	err = h.pgClient.QueryRow(r.Context(), `
		SELECT service.update_application_status($1, $2, $3)
	`, email, applicationID, newStatus,
	).Scan(&success)

	if err != nil {
		fmt.Printf("Error calling update_application_status: %v\n", err)
		http.Error(w, "Error editing application status", http.StatusInternalServerError)
		return
	}

	if !success {
		fmt.Println("Error: update_application_status failed")
		http.Error(w, "Error editing application status", http.StatusInternalServerError)
		return
	}

	log.Println("Application status edited")
	log.Println("-----------------")

	w.WriteHeader(http.StatusOK)
}

// does not invalidate cache because this is just a metadata update
func (h *Handler) EditApplication(w http.ResponseWriter, r *http.Request) {
	log.Println("[*] EditApplication [*]")
	log.Println("-----------------")

	email, err := auth.IsAuthenticated(r)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Println("User authenticated")

	// extract json from request body
	var editApplicationRequest EditApplicationRequest
	err = json.NewDecoder(r.Body).Decode(&editApplicationRequest)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	applicationID := editApplicationRequest.ID
	role := editApplicationRequest.Role
	company := editApplicationRequest.Company
	location := editApplicationRequest.Location
	link := editApplicationRequest.Link


	// update application in Postgres
	var success bool

	err = h.pgClient.QueryRow(r.Context(), `
		SELECT service.edit_application($1, $2, $3, $4, $5, $6)
	`, email, applicationID, company, role, link, location,
	).Scan(&success)
	if err != nil {
		fmt.Printf("Error calling edit_application: %v\n", err)
		http.Error(w, "Error editing application", http.StatusInternalServerError)
		return
	}

	if !success {
		fmt.Println("Error: edit_application failed")
		http.Error(w, "Error editing application", http.StatusInternalServerError)
		return
	}

	log.Println("Application edited")
	log.Println("-----------------")

	w.WriteHeader(http.StatusOK)
}

// don't invalidate cache because this does not change the applied date
func (h *Handler) RevertStatus(w http.ResponseWriter, r *http.Request) {
	log.Println("[*] RevertStatus [*]")
	log.Println("-----------------")

	email, err := auth.IsAuthenticated(r)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Println("User authenticated")

	// extract the jobID requested to revert, revert it in database and send to PubSub
	// NOTE: BigQuery is not optimized for single-row deletes, so we should simply set a 
	// "reverted" flag to the most recent event and ensure analytics only reads non-flagged events
	// DeleteApplication and DeleteUser are okay because they are multi-row deletes and much less frequent
	var revertApplicationStatusRequest RevertApplicationStatusRequest
	err = json.NewDecoder(r.Body).Decode(&revertApplicationStatusRequest)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	operationID := revertApplicationStatusRequest.OperationID
	jobID := revertApplicationStatusRequest.ID

	var newStatus sql.NullString

	err = h.pgClient.QueryRow(r.Context(), `
		SELECT service.revert_operation($1, $2, $3)
	`, email, jobID, operationID).Scan(&newStatus)
	if err != nil {
		fmt.Printf("Error calling revert_operation: %v\n", err)
		http.Error(w, "Error reverting status", http.StatusInternalServerError)
		return
	}

	if !newStatus.Valid {
		fmt.Println("Error: revert_operation failed")
		http.Error(w, "Error reverting status", http.StatusInternalServerError)
		return
	}

	log.Println("Application status reverted")
	log.Println("-----------------")

	// return the latest status for optimistc UI
	response := map[string]interface{}{
		"status": newStatus.String,
	}

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	log.Println("[*] DeleteUser [*]")
	log.Println("-----------------")

	email, err := auth.IsAuthenticated(r)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Println("User authenticated")

	var success bool
	err = h.pgClient.QueryRow(r.Context(), `
		SELECT service.delete_user($1)
	`, email).Scan(&success)
	if err != nil {
		fmt.Printf("Error calling delete_user: %v\n", err)
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	if !success {
		fmt.Println("Error: delete_user failed")
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}


	log.Println("User deleted")
	log.Println("-----------------")

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetApplicationTimeline(w http.ResponseWriter, r *http.Request) {
	log.Println("[*] GetApplicationTimeline [*]")
	log.Println("-----------------")

	email, err := auth.IsAuthenticated(r)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Println("User authenticated")

	var getApplicationTimelineRequest ApplicationTimelineRequest
	err = json.NewDecoder(r.Body).Decode(&getApplicationTimelineRequest)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	jobID := getApplicationTimelineRequest.ID

	rows, err := h.pgClient.Query(r.Context(), `
		SELECT * from service.get_application_timeline($1, $2)
	`, email, jobID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Error getting application timeline", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	timeline := make([]Operation, 0)
	for rows.Next() {
		var (
			operationID string
			eventTime   time.Time
			status      string
			operation   string
		)
		err := rows.Scan(
			&operationID,
			&eventTime,
			&status,
			&operation,
		)
		if err != nil {
			fmt.Printf("Error scanning row: %v\n", err)
			http.Error(w, "Error scanning row", http.StatusInternalServerError)
			return
		}
		timeline = append(timeline, Operation{
			OperationID: operationID,
			Operation:   operation,
			Status:      status,
			EventTime:   eventTime.Local(),	// postgres stores UTC, convert local
		})
	}
	if err := rows.Err(); err != nil {
		fmt.Printf("Error iterating rows: %v\n", err)
		http.Error(w, "Error iterating rows", http.StatusInternalServerError)
		return
	}

	log.Println("Timeline extracted")
	log.Println("-----------------")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(timeline)
}

func (h *Handler) extractAnalytics(ctx context.Context, email string) (map[string]interface{}, error) {
	res := map[string]interface{}{
		"email": email,
	}

	// im sorry bruh this is the only way to do it lol
	var (
		applicationsCount, appliedCount, ghostedCount, rejectedCount, screenCount, interviewingCount,
		offerCount, applicationVelocity, applicationVelocityTrend, resumeEffectiveness,
		resumeEffectivenessTrend, interviewEffectiveness, interviewEffectivenessTrend,
		avgResponseTime, prevAvgResponseTime, avgResponseTimeTrend sql.NullInt64
	) 

	var yearlyTrends []byte

	err := h.pgClient.QueryRow(ctx, "SELECT * FROM service.profile($1)", email).Scan(
		&applicationsCount, &appliedCount, &ghostedCount, &rejectedCount, &screenCount,
		&interviewingCount, &offerCount, &applicationVelocity, &applicationVelocityTrend,
		&resumeEffectiveness, &resumeEffectivenessTrend, &interviewEffectiveness,
		&interviewEffectivenessTrend, &avgResponseTime, &prevAvgResponseTime, &avgResponseTimeTrend,
		&yearlyTrends,
	)
	if err != nil {
		// if no profile found (pgx.ErrNoRows), umm how tf did u even authenticate this call
		// anyways just return nothing cause this should never happen
		fmt.Printf("Error querying profile: %v\n", err)
		return nil, err
	}

	// set nullable int values to nil if not valid
	setNullableInt := func(m map[string]interface{}, key string, val sql.NullInt64) {
		if val.Valid {
			m[key] = val.Int64
		} else {
			m[key] = nil
		}
	
	}

	// so fucking ugly sorry
	setNullableInt(res, "applications_count", applicationsCount)
	setNullableInt(res, "applied_count", appliedCount)
	setNullableInt(res, "ghosted_count", ghostedCount)
	setNullableInt(res, "rejected_count", rejectedCount)
	setNullableInt(res, "screen_count", screenCount)
	setNullableInt(res, "interviewing_count", interviewingCount)
	setNullableInt(res, "offer_count", offerCount)
	setNullableInt(res, "application_velocity", applicationVelocity)
	setNullableInt(res, "application_velocity_trend", applicationVelocityTrend)
	setNullableInt(res, "resume_effectiveness", resumeEffectiveness)
	setNullableInt(res, "resume_effectiveness_trend", resumeEffectivenessTrend)
	setNullableInt(res, "interview_effectiveness", interviewEffectiveness)
	setNullableInt(res, "interview_effectiveness_trend", interviewEffectivenessTrend)
	setNullableInt(res, "avg_response_time", avgResponseTime)
	setNullableInt(res, "prev_avg_response_time", prevAvgResponseTime)
	setNullableInt(res, "avg_response_time_trend", avgResponseTimeTrend)

	// parse JSONB yearly trends
    if len(yearlyTrends) > 0 {
        var trends map[string]interface{}
        if err := json.Unmarshal(yearlyTrends, &trends); err == nil {
            res["yearly_trends"] = trends
        } else {
            res["yearly_trends"] = map[string]interface{}{}
        }
    } else {
        res["yearly_trends"] = map[string]interface{}{}
    }

    return res, nil
}

// note that prevPageBoundary is unix and we always store timezone-independent in Postgres
// so this is safe to use :))))))) i tested btw it works :))))
func (h *Handler) filterQueryText(email, queryText string, ctx context.Context, hitsPerPageInt, page int, prevPageBoundary *time.Time) ([]SearchResponse, int64, error) {
    var applications []SearchResponse
    var rows pgx.Rows
    var err error

    if prevPageBoundary != nil {
        // use keyset pagiantion
        rows, err = h.pgClient.Query(ctx, `
            SELECT * FROM service.get_user_applications($1, $2, $3, $4, $5)
        `, email, queryText, hitsPerPageInt, 0, prevPageBoundary)
    } else {
        // use offset pagination
        rows, err = h.pgClient.Query(ctx, `
            SELECT * FROM service.get_user_applications($1, $2, $3, $4, $5)
        `, email, queryText, hitsPerPageInt, page*hitsPerPageInt, nil)
    }
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return nil, 0, err
	}
	defer rows.Close()

	var totalHits sql.NullInt64

	for rows.Next() {
		var (
			applicationID sql.NullString
			email         sql.NullString	// not used in the response, but scanned
			appliedDate   time.Time
			operation     sql.NullString	// might be useful if needed; not used below
			appStatus     sql.NullString
			company       sql.NullString
			title         sql.NullString	// we treat title as the Role
			link          sql.NullString
			locations     sql.NullString	// we treat locations as Location
		)
		err := rows.Scan(
			&applicationID,
			&email,
			&appliedDate,
			&operation,
			&appStatus,
			&company,
			&title,
			&link,
			&locations,
			&totalHits,
		)
		if err != nil {
			// handle scan error
			fmt.Printf("Error scanning row: %v\n", err)
			return nil, 0, err
		}

		appStatusStr := ""
		if appStatus.Valid {
			appStatusStr = appStatus.String
		}
		
		application := SearchResponse{
			ID:          applicationID.String,
			Role:        title.String,
			Company:     company.String,
			Location:    locations.String,
			AppliedDate: appliedDate.UnixMilli(),
			Status:      ApplicationStatus(appStatusStr),	// ApplicationStatus() is basically enum wrapper
			Link:        link.String,
		}
		
		applications = append(applications, application)
	}

	var hits int64 = 0
	if totalHits.Valid {
		hits = int64(totalHits.Int64)
	}

	if err := rows.Err(); err != nil {
		fmt.Printf("Error iterating rows: %v\n", err)
		return nil, 0, err
	}

	return applications, hits, nil
}

// for every data mutation, change cache version
func (h *Handler) invalidateUserCache(ctx context.Context, email string) {
	// increment the cache version for this user
	versionKey := fmt.Sprintf("user:%s:cache_version", email)
	h.redisClient.Incr(ctx, versionKey)
}