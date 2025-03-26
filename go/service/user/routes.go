package user

// this file contains the HTTP handlers for the user service
// it contains the following handlers:
// (R) - Dashboard: queries Algolia for applications based on search query
// (R) - Profile: (for now) returns simply email and app count; once we figure out what kind of data analytics we want to show, it will be updated
// (C) - AddApplication: adds an application to Firestore and publishes a message to PubSub
// (D) - DeleteApplication: deletes an application from Firestore and publishes a message to PubSub
// (U) - EditStatus: edits the status of an application in Firestore and publishes a message to PubSub
// (U) - EditApplication: edits an application in Firestore and publishes a message to PubSub
// (D) - DeleteUser: deletes a user from Firestore and publishes a message to PubSub to delete all applications from Algolia
// this file contains the following utility functions:
// - deleteUserFromFirestore: deletes a user from Firestore, including all applications
// - publishMessage: publishes a message to PubSub with publish and connection retries
//     (relies on utils.PublishWithRetry)
// NOTE: all CRUD operations (AddApplication, DeleteApplication, EditStatus, EditApplication, DeleteUser) are idempotent
//       and can be retried without side effects. This is why there is no timestamping or versioning.
// NOTE: all CRUD operations are NOT commutative. We rely on an optimistic but strong consistency model. So,
//	 	 every user request is fulfilled to DB, but in the rare event of publishing failure, we can quickly revert.
//		 Actually, this does not add any latency because we have frontend send previous state in the delete or edit request
//       so there's no need to query the DB for the previous state.
// Q: why not use event sourcing?
// A: all that matters is latest state; rebuilding history is not necessary. HOWEVER, compliance and auditing
//    may require event sourcing in the future if this project takes off
// Q: why is EditStatus and EditApplication separate?
// A: EditStatus has a different UI flow and separate actions, mainly because a user will very rarely edit
//    fields such as role, company, applied date, etc. but will frequently edit status. This separation
//    allows for a more optimized UI flow and better user experience

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/copium-dev/copium/go/service/auth"
	"github.com/copium-dev/copium/go/service/user/userutils"
	"github.com/copium-dev/copium/go/utils"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/bigquery"
	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
)

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
	Applications []AlgoliaResponse `json:"applications"`
	TotalPages   int               `json:"totalPages"`
	CurrentPage  int               `json:"currentPage"`
}

type AlgoliaResponse struct {
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
	OldStatus   ApplicationStatus `json:"oldStatus"`
	AppliedDate int64  `json:"appliedDate"`
}

// edit application does not include status because status is edited separately
type EditApplicationRequest struct {
	ID             string `json:"id"`
	Role           string `json:"role"`
	Company        string `json:"company"`
	Location       string `json:"location"`
	Link           string `json:"link"`
	OldRole        string `json:"oldRole"`
	OldCompany     string `json:"oldCompany"`
	OldLocation    string `json:"oldLocation"`
	OldLink        string `json:"oldLink"`
	Status         ApplicationStatus `json:status`
}

type RevertApplicationStatusRequest struct {
	OperationID string `json:"operationID"`
	ID       string `json:"id"`
}

type ApplicationTimelineRequest struct {
	ID string `json:"id"`
}

type Handler struct {
	FirestoreClient *firestore.Client
	algoliaClient   *search.APIClient
	bigQueryClient *bigquery.Client
	pubsubTopic     *pubsub.Topic
	orderingKey     string
}

func NewHandler(
	firestoreClient *firestore.Client,
	algoliaClient *search.APIClient,
	bigQueryClient *bigquery.Client,
	pubsubTopic *pubsub.Topic,
	orderingKey string,
) *Handler {
	return &Handler{
		FirestoreClient: firestoreClient,
		algoliaClient:   algoliaClient,
		bigQueryClient: bigQueryClient,
		pubsubTopic:     pubsubTopic,
		orderingKey:     orderingKey,
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

	// get user's applications count
	doc, err := h.FirestoreClient.Collection("users").Doc(email).Get(r.Context())
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Error retrieving user data", http.StatusInternalServerError)
		return
	}

	userData := doc.Data()

	applicationsCount := int64(0)
	if countVal, exists := userData["applicationsCount"]; exists && countVal != nil {
		if count, ok := countVal.(int64); ok {
			applicationsCount = count
		}
	}

	response := map[string]interface{}{
		"email":             email,
		"applicationsCount": applicationsCount,
	}

	analyticsFields := []string{
		"application_velocity_trend", "application_velocity",
		"resume_effectiveness_trend", "resume_effectiveness",
		"interview_effectiveness_trend", "interview_effectiveness",
		"avg_response_time_trend", "avg_response_time",
		"monthly_trends", "rejected_count", "ghosted_count", "applied_count",
		"screen_count", "interviewing_count", "offer_count", "last_updated",
	}

	// loop over each field and add to response if it exists
	for _, field := range analyticsFields {
		if val, exists := doc.Data()[field]; exists {
			response[field] = val
		}
	}

	log.Println("Profile data extracted")
	log.Println("-----------------")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// queries algolia for applications based on search query
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

	// 1. extract search query from request and parse
	queryText, filtersString, err := userutils.ParseQuery(r)
	// any invalid query params will return a 400 error
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Error parsing query", http.StatusBadRequest)
		return
	}

	log.Println("Filters parsed", filtersString)

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

	hitsPerPage := r.URL.Query().Get("hits")
	hitsPerPageInt := 10 // Default value

	if hitsPerPage != "" {
		parsed, err := strconv.Atoi(hitsPerPage)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			http.Error(w, "Error parsing hitsPerPage", http.StatusBadRequest)
			return
		}
		hitsPerPageInt = parsed
	}

	if hitsPerPageInt < 10 {
		hitsPerPageInt = 10
	} else if hitsPerPageInt > 18 {
		hitsPerPageInt = 18
	}

	log.Println("Hits per page requested:", hitsPerPageInt)

	// 2. build a search params object
	searchParamsObject := &search.SearchParamsObject{
		Facets:       []string{"email"},
		FacetFilters: &search.FacetFilters{String: utils.StringPtr("email:" + email)},
		HitsPerPage:  utils.IntPtr(int32(hitsPerPageInt)),
		Filters:      utils.StringPtr(filtersString),
		Page:         utils.IntPtr(int32(page)),
	}

	// set free text query if present
	if queryText != "" {
		searchParamsObject.Query = utils.StringPtr(queryText)
		log.Println("Free text query text extracted: ", queryText)
	}

	searchParams := &search.SearchParams{
		SearchParamsObject: searchParamsObject,
	}

	// 3. query Algolia with the search params
	response, err := h.algoliaClient.SearchSingleIndex(
		h.algoliaClient.NewApiSearchSingleIndexRequest("users").WithSearchParams(searchParams),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Error querying Algolia", http.StatusInternalServerError)
		return
	}

	// 4. extract hits from response
	var applications []AlgoliaResponse

	// 4a. marshal the raw hits into JSON bytes.
	hitsBytes, err := json.Marshal(response.Hits)
	if err != nil {
		fmt.Printf("Error marshaling hits: %v\n", err)
		http.Error(w, "Error processing hits", http.StatusInternalServerError)
		return
	}

	// 4b. unmarshal the JSON bytes into AlgoliaResponse slice.
	err = json.Unmarshal(hitsBytes, &applications)
	if err != nil {
		fmt.Printf("Error unmarshaling hits: %v\n", err)
		http.Error(w, "Error processing applications", http.StatusInternalServerError)
		return
	}

	log.Println("Applications extracted:", applications)
	log.Println("-----------------")

	// create response object pagination info
	responseObject := DashboardResponse{
		Applications: applications,
		TotalPages:   userutils.CalculateTotalPages(int(*response.NbHits), hitsPerPageInt),
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

	// extract json from request body
	var addApplicationRequest AddApplicationRequest
	err = json.NewDecoder(r.Body).Decode(&addApplicationRequest)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	// add application to Firestore using users/{email} where jobs is a document within the user's collection
	doc, _, err := h.FirestoreClient.Collection("users").Doc(email).Collection("applications").Add(r.Context(), map[string]interface{}{
		"role":        addApplicationRequest.Role,
		"company":     addApplicationRequest.Company,
		"location":    addApplicationRequest.Location,
		"appliedDate": addApplicationRequest.AppliedDate,
		"status":      addApplicationRequest.Status,
		"link":        addApplicationRequest.Link,
	})
	if err != nil {
		fmt.Printf("Error adding application: %v\n", err)
		http.Error(w, "Error adding application", http.StatusInternalServerError)
		return
	}

	log.Println("Application added")

	message := map[string]interface{}{
		"operation":   "add",
		"email":       email,
		"appliedDate": addApplicationRequest.AppliedDate,
		"company":     addApplicationRequest.Company,
		"link":        addApplicationRequest.Link,
		"location":    addApplicationRequest.Location,
		"role":        addApplicationRequest.Role,
		"status":      addApplicationRequest.Status,
		"timestamp":   time.Now().Add(12 * time.Hour).Unix(),
		"objectID":    doc.ID,
	}

	err = h.publishMessage(message)
	if err != nil {
		// delete added application if publish fails
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = h.FirestoreClient.Collection("users").Doc(email).Collection("applications").Doc(doc.ID).Delete(ctx)
		if err != nil {
			fmt.Printf("Error deleting application: %v\n", err)
			http.Error(w, "Error reverting application add", http.StatusInternalServerError)
		}
		log.Println("AddApplication reverted because of publish failure")
		http.Error(w, "Error publishing message", http.StatusInternalServerError)
		return
	}

	// to reduce amount of reads in this single request, increment applicationCount AFTER verifying publish success
	// this means we don't have to revert applicationCount on top of reverting the add operation. this DOES
	// introduce small window of inconsistency but this is reducing costs and reducing complexity
	userDoc := h.FirestoreClient.Collection("users").Doc(email)
	_, err = userDoc.Update(r.Context(), []firestore.Update{
		{Path: "applicationsCount", Value: firestore.Increment(1)},
		{Path: "applied_count", Value: firestore.Increment(1)},
	})
	if err != nil {
		fmt.Printf("Error updating applications count: %v\n", err)
		http.Error(w, "Error updating application count", http.StatusInternalServerError)
		return
	}

	log.Println("Applications count updated, added by 1")
	log.Println("DB and PubSub operations success, returning ID for eager loading")

	// return doc.ID to user for eager loading
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"objectID": doc.ID,
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

	// delete application from Firestore
	_, err = h.FirestoreClient.Collection("users").Doc(email).Collection("applications").Doc(applicationID).Delete(r.Context())
	if err != nil {
		fmt.Printf("Error deleting application: %v\n", err)
		http.Error(w, "Error deleting application", http.StatusInternalServerError)
		return
	}

	log.Println("Application deleted")

	message := map[string]interface{}{
		"operation": "delete",
		"email":     email,
		"objectID":  applicationID,
	}

	err = h.publishMessage(message)
	if err != nil {
		// revert status if publish fails
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = h.FirestoreClient.Collection("users").Doc(email).Collection("applications").Doc(applicationID).Set(ctx, map[string]interface{}{
			"role":        deleteApplicationRequest.Role,
			"company":     deleteApplicationRequest.Company,
			"location":    deleteApplicationRequest.Location,
			"appliedDate": deleteApplicationRequest.AppliedDate,
			"status":      deleteApplicationRequest.Status,
			"link":        deleteApplicationRequest.Link,
		})
		if err != nil {
			fmt.Printf("Error reverting application: %v\n", err)
			http.Error(w, "Error reverting application delete", http.StatusInternalServerError)
			return
		}
		log.Println("DeleteApplication reverted because of publish failure")
		http.Error(w, "Error publishing message", http.StatusInternalServerError)
		return
	}

	// to reduce amount of reads in this single request, decrement applicationCount AFTER verifying publish success
	// this means we don't have to revert applicationCount on top of reverting the delete operation. this DOES
	// introduce small window of inconsistency but this is reducing costs and reducing complexity
	userDoc := h.FirestoreClient.Collection("users").Doc(email)
	_, err = userDoc.Update(r.Context(), []firestore.Update{
		{Path: "applicationsCount", Value: firestore.Increment(-1)},
		{Path: fmt.Sprintf("%s_count", strings.ToLower(string(deleteApplicationRequest.Status))), Value: firestore.Increment(-1)},
	})
	if err != nil {
		fmt.Printf("Error updating applications count: %v\n", err)
		http.Error(w, "Error updating application count", http.StatusInternalServerError)
	}

	log.Println("Applications count updated, decremented by 1")
	log.Println("DB and PubSub operations success, returning success for eager loading")

	w.WriteHeader(http.StatusOK)
}

// NOTE: eager loading is not necessary here because frontend already assumes success
// if any error, a refresh would show the correct (reverted) state
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

	applicationID := EditApplicationStatusRequest.ID
	newStatus := EditApplicationStatusRequest.Status
	appliedDate := EditApplicationStatusRequest.AppliedDate

	// before doing DB updates check if we even need to update
	if newStatus == EditApplicationStatusRequest.OldStatus {
		log.Println("No status change, returning success")
		w.WriteHeader(http.StatusOK)
		return
	}

	_, err = h.FirestoreClient.Collection("users").Doc(email).Collection("applications").Doc(applicationID).Update(r.Context(), []firestore.Update{
		{
			Path:  "status",
			Value: newStatus,
		},
	})
	if err != nil {
		fmt.Printf("Error editing application: %v\n", err)
		http.Error(w, "Error editing application", http.StatusInternalServerError)
		return
	}

	log.Println("Application status edited")

	message := map[string]interface{}{
		"operation":   "editStatus",
		"email":       email,
		"objectID":    applicationID,
		"status":      newStatus,
		"appliedDate": appliedDate,	// just to satisfy BigQuery schema
		// since appliedDate is always using noon as the time, we need to ensure
		// that the timestamp sent to PubSub is always at or after noon. this is because
		// a user can edit status of an application at 11:59 AM and the appliedDate is 12:00 PM
		// so this will cause response time metrics to be incorrect
		// so, simply add 12 hours to guarantee it's always at or after noon
		"timestamp": time.Now().Add(12 * time.Hour).Unix(),
	}

	err = h.publishMessage(message)
	if err != nil {
		// revert status if publish fails
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = h.FirestoreClient.Collection("users").Doc(email).Collection("applications").Doc(applicationID).Update(ctx, []firestore.Update{
			{
				Path:  "status",
				Value: EditApplicationStatusRequest.OldStatus,
			},
		})
		if err != nil {
			fmt.Printf("Error reverting status: %v\n", err)
			http.Error(w, "Error reverting status", http.StatusInternalServerError)
			return
		}
		log.Println("EditStatus reverted because of publish failure")
		http.Error(w, "Error publishing message", http.StatusInternalServerError)
		return
	}

	log.Println("Status edited")

	// to reduce amount of reads in this single request, update status count AFTER verifying publish success
	// this means we don't have to revert status count on top of reverting the edit operation. this DOES
	// introduce small window of inconsistency but this is reducing costs and reducing complexity
	userDoc := h.FirestoreClient.Collection("users").Doc(email)
	_, err = userDoc.Update(r.Context(), []firestore.Update{
		{Path: fmt.Sprintf("%s_count", strings.ToLower(string(newStatus))), Value: firestore.Increment(1)},
		{Path: fmt.Sprintf("%s_count", strings.ToLower(string(EditApplicationStatusRequest.OldStatus))), Value: firestore.Increment(-1)},
	})
	if err != nil {
		fmt.Printf("Error updating status count: %v\n", err)
		http.Error(w, "Error updating status count", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

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

	// 1. check if we need to update at all but also
	// 2. populate a changedFields map so that we only update what we have to in Firestore
	// unfortunately though Algolia **does** require all fields to be sent regardless
	// so this is just a little optimization on the Firestore side
	// NOTE: the EditApplicationRequest struct's fields are still required for the publish message
	// and possible rollbacks, so frontend cannot make the optimization of what to send
	changedFields := make(map[string]interface{}, 0)

	// no loops or function calls to reduce memory overhead, big ugly if statements
	if editApplicationRequest.Role != editApplicationRequest.OldRole {
		changedFields["role"] = editApplicationRequest.Role
	}
	if editApplicationRequest.Company != editApplicationRequest.OldCompany {
		changedFields["company"] = editApplicationRequest.Company
	}
	if editApplicationRequest.Location != editApplicationRequest.OldLocation {
		changedFields["location"] = editApplicationRequest.Location
	}
	if editApplicationRequest.Link != editApplicationRequest.OldLink {
		changedFields["link"] = editApplicationRequest.Link
	}
	
	if len(changedFields) == 0 {
		log.Println("No application change, returning success")
		w.WriteHeader(http.StatusOK)
		return
	}

	// create updates array for Firestore
	updates := make([]firestore.Update, len(changedFields))
	i := 0
	for key, value := range changedFields {
		updates[i] = firestore.Update{Path: key, Value: value}
		i++
	}

	_, err = h.FirestoreClient.Collection("users").Doc(email).Collection("applications").Doc(applicationID).Update(r.Context(), updates)
	if err != nil {
		fmt.Printf("Error editing application: %v\n", err)
		http.Error(w, "Error editing application", http.StatusInternalServerError)
		return
	}

	log.Println("Application edited")

	// bigquery does nothing on application edits, only status changes
	// this is why we need a diffentiating operation for application edits
	// is this wasted data transfer? yea... but its not a lot of data and
	// not worth setting up different messaging pipeline when just one operation is not supported by BigQuery
	message := map[string]interface{}{
		"operation":   "editApplication",
		"email":       email,
		"company":     editApplicationRequest.Company,
		"link":        editApplicationRequest.Link,
		"location":    editApplicationRequest.Location,
		"role":        editApplicationRequest.Role,
		"objectID":    applicationID,
		"timestamp":   time.Now().Add(12 * time.Hour).Unix(),
	}

	err = h.publishMessage(message)
	if err != nil {
		// revert status if publish fails
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = h.FirestoreClient.Collection("users").Doc(email).Collection("applications").Doc(applicationID).Update(ctx, []firestore.Update{
			{Path: "role", Value: editApplicationRequest.OldRole},
			{Path: "company", Value: editApplicationRequest.OldCompany},
			{Path: "location", Value: editApplicationRequest.OldLocation},
			{Path: "link", Value: editApplicationRequest.OldLink},
		})
		if err != nil {
			fmt.Printf("Error reverting application: %v\n", err)
			http.Error(w, "Error reverting application edit", http.StatusInternalServerError)
			return
		}
		log.Println("EditApplication reverted because of publish failure")
		http.Error(w, "Error publishing message", http.StatusInternalServerError)
		return
	}

	log.Println("Application edited")
	log.Println("DB and PubSub operations success, returning success for eager loading")

	w.WriteHeader(http.StatusOK)
}

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

	// two cases:
	// 1: operation is most recent (get from BigQuery); we have to update Algolia and Firestore to the previous status
	//	  compensating transaction is needed here because latest state may have changed
	// 2: operation not most recent; simply flag the operation as "reverted" in BigQuery. Algolia and Firestore still have most recent status
	//	  compensating transaction is not needed here because latest state never changed
	operationID := revertApplicationStatusRequest.OperationID
	jobID := revertApplicationStatusRequest.ID

	// get the two max event times; the first is to determine if case 2, the second is to revert if case 1
	q := h.bigQueryClient.Query(`
		SELECT operationID, status, event_time  
		FROM applications_data.applications
		WHERE email = @email
		AND jobID = @jobID
		AND operation != 'revert'  -- skip any previously reverted operations
		ORDER BY event_time DESC
		LIMIT 2
	`)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "email", Value: email},
		{Name: "jobID", Value: jobID},
	}

	job, err := q.Run(r.Context())
	if err != nil {
		fmt.Printf("Error getting max event time: %v\n", err)
		http.Error(w, "Error reverting status", http.StatusInternalServerError)
		return
	}

	_, err = job.Wait(r.Context())
	if err != nil {
		fmt.Printf("Error getting max event time: %v\n", err)
		http.Error(w, "Error reverting status", http.StatusInternalServerError)
		return
	}

	it, err := job.Read(r.Context())
	if err != nil {
		fmt.Printf("Error reading query results: %v\n", err)
		http.Error(w, "Error reverting status", http.StatusInternalServerError)
		return
	}

	var latestOperationID string
	var currStatus string	// just in case we need to rollback
	var secondLatestOperationID string
	var prevStatus string
	rowCount := 0

	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Printf("Error getting max event time: %v\n", err)
			http.Error(w, "Error reverting status", http.StatusInternalServerError)
			return
		}

		if rowCount == 0 {
			latestOperationID = row[0].(string)    
			currStatus = row[1].(string)           
		} else if rowCount == 1 {
			secondLatestOperationID = row[0].(string)   
			prevStatus = row[1].(string)          
		}

		rowCount++
	}

	// if there's only one operation (or none), there's nothing to revert to
	if rowCount < 2 {

		http.Error(w, "Not enough operations to revert", http.StatusBadRequest)
		return
	}

	// ensure operationID is valid
	operationExists := false
	if latestOperationID == operationID || secondLatestOperationID == operationID {
		operationExists = true
	}

	if !operationExists {
		http.Error(w, "Operation not found", http.StatusNotFound)
		return
	}

	var operation string

	if latestOperationID != operationID {
		// case 2: flag as reverted in BQ. Algolia and Firestore are already up to date
		operation = "revert"
		fmt.Println("Case 2: Reverting deeper in history -- only BigQuery needs to be updated")
	} else {
		// case 1: Firestore and Algolia need to be updated to previous status (secondLatestOperation)
		operation = "revertLatest"
		fmt.Println("Case 1: Reverting most recent operation -- Firestore and Algolia need to be updated as well")
	}

	_, err = h.FirestoreClient.Collection("users").Doc(email).Collection("applications").Doc(jobID).Update(r.Context(), []firestore.Update{
		{Path: "status", Value: prevStatus},
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Error reverting status", http.StatusInternalServerError)
		return
	}

	log.Println("Status reverted")

	// timestamp not needed because it will set the most recent event to "reverted"
	// since we already prevent duplicate status updates and have a safeguard against double reverts,
	// simply editing the most recent timestamp is perfectly fine
	message := map[string]interface{}{
		"operation": operation,
		"email":     email,
		"objectID":  jobID,
		"operationID": operationID,
		"status":    prevStatus,
	}

	err = h.publishMessage(message)
	// revertLatest is a special case -- need to revert Firestore status if publish fails
	if err != nil && operation == "revertLatest" {
		// revert status if publish fails
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = h.FirestoreClient.Collection("users").Doc(email).Collection("applications").Doc(jobID).Update(ctx, []firestore.Update{
			{Path: "status", Value: currStatus},
		})
		if err != nil {
			fmt.Printf("Error reverting application: %v\n", err)
			http.Error(w, "Error reverting application edit", http.StatusInternalServerError)
			return
		}
		log.Println("RevertStatus reverted because of publish failure")
		http.Error(w, "Error publishing message", http.StatusInternalServerError)
		return
	}

	log.Println("RevertStatus success")
	// note: frontend has no way to access previous state
	// so, when user clicks revert we need to refresh on ok instead of optimistic UI
	w.WriteHeader(http.StatusOK)
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

	// send to algolia to delete all applications associated with this user
	message := map[string]interface{}{
		"operation": "userDelete",
		"email":     email,
	}

	err = h.publishMessage(message)
	if err != nil {
		fmt.Printf("Error publishing message: %v\n", err)
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	// since we can't exactly revert a user deletion, we will delete only if publish is successful
	err = h.deleteUserFromFirestore(email, 10)
	if err != nil {
		fmt.Printf("Error deleting user: %v\n", err)
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	log.Println("User deleted")

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

	q := h.bigQueryClient.Query(`
		SELECT operationID, operation, status, event_time
		FROM applications_data.applications
		WHERE email = @email
		AND jobID = @jobID
		AND operation != 'revert'
		ORDER BY event_time DESC
	`)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "email", Value: email},
		{Name: "jobID", Value: jobID},
	}

	job, err := q.Run(r.Context())
	if err != nil {
		fmt.Printf("Error getting timeline: %v\n", err)
		http.Error(w, "Error getting timeline", http.StatusInternalServerError)
		return
	}

	status, err := job.Wait(r.Context())
	if err != nil {
		fmt.Printf("Error getting timeline: %v\n", err)
		http.Error(w, "Error getting timeline", http.StatusInternalServerError)
		return
	}
	if err := status.Err(); err != nil {
		fmt.Printf("Job completed with erorr: %v\n", err)
		http.Error(w, "Job completed with error", http.StatusInternalServerError)
		return
	}


	it, err := job.Read(r.Context())
	if err != nil {
		fmt.Printf("Error reading timeline: %v\n", err)
		http.Error(w, "Error reading timeline", http.StatusInternalServerError)
		return
	}

	type Row struct {
		OperationID string `bigquery:"operationID"`
		Operation   string `bigquery:"operation"`
		Status      string `bigquery:"status"`
		EventTime   time.Time `bigquery:"event_time"`
	}

	var rows []Operation

	for {
		var row Row
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Printf("Error iterating timeline: %v\n", err)
			http.Error(w, "Error iterating timeline", http.StatusInternalServerError)
			return
		}
		rows = append(rows, Operation{
			OperationID: row.OperationID,
			Operation:   row.Operation,
			Status:      row.Status,
			EventTime:   row.EventTime,
		})
	}

	log.Println("Timeline extracted")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rows)
}

// publishes to both algolia and bigquery topics
func (h *Handler) publishMessage(message map[string]interface{}) error {
	// new context here -- message should be published regardless of context cancellation
	// for consistency enforcement. use 10 second timeout to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	messageBody, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("Error marshaling message: %v\n", err)
		return err
	}

	// hold the publish result. this is necessary for
	// our strong consistency model
	var result *pubsub.PublishResult

	// attempt to publish message (algolia and bigquery both subscribe to this topic)
	r := h.pubsubTopic.Publish(ctx, &pubsub.Message{
		Data:        messageBody,
		OrderingKey: h.orderingKey,
	})

	// if message publish fails, propagate an error to revert Firestore operation
	result = r
	id, err := result.Get(ctx)
	if err != nil {
		fmt.Printf("Error publishing message: %v\n", err)
		return err
	}

	fmt.Printf("Published message with ID: %s\n", id)

	return nil
}

// Firestore does not delete subcollections automatically
// so, delete all documents in users/{email}/applications
// then, delete users/{email}
func (h *Handler) deleteUserFromFirestore(email string, batchSize int) error {
	// a user might just close the tab after running delete, so we need to ensure
	// that the context is not cancelled and the delete still goes through
	ctx := context.Background()

	// delete subcollection FIRST (just applications)
	applicationsCollection := h.FirestoreClient.Collection("users").Doc(email).Collection("applications")
	bulkWriter := h.FirestoreClient.BulkWriter(ctx)

	// for each batch...
	for {
		iter := applicationsCollection.Limit(batchSize).Documents(ctx)
		numDeleted := 0

		// for each document...
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return fmt.Errorf("Failed to iterate: %v", err)
			}

			bulkWriter.Delete(doc.Ref)
			numDeleted++
		}

		if numDeleted == 0 {
			bulkWriter.End()
			break
		}

		bulkWriter.Flush()
	}

	fmt.Println("Applications subcollection deleted for user", email)

	// delete user document
	_, err := h.FirestoreClient.Collection("users").Doc(email).Delete(ctx)
	if err != nil {
		return fmt.Errorf("Failed to delete user document: %v", err)
	}

	fmt.Println("User document deleted for user", email)

	return nil
}