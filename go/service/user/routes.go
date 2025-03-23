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
//     (relies on utils.PublishWithRetry and retryRabbitConnectionAndRetryPublish)
// - retryRabbitConnectionAndRetryPublish: re-establishes connection to PubSub and retries publishing a message
//	   (relies on utils.RetryRabbitConnection and utils.PublishWithRetry)
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
	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
)

type AddApplicationRequest struct {
	Role        string `json:"role"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	AppliedDate int64  `json:"appliedDate"`
	Status      string `json:"status"`
	Link        string `json:"link"`
}

type Application struct {
	ID          string `json:"id"`
	Role        string `json:"role"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	AppliedDate int64  `json:"appliedDate"`
	Status      string `json:"status"`
	Link        string `json:"link"`
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
	Status      string `json:"status"`
	Link        string `json:"link"`
}

type DeleteApplicationRequest struct {
	ID          string `json:"id"`
	Role        string `json:"role"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	AppliedDate int64  `json:"appliedDate"`
	Status      string `json:"status"`
	Link        string `json:"link"`
}

type EditApplicationStatusRequest struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	OldStatus   string `json:"oldStatus"`
	AppliedDate int64  `json:"appliedDate"`
}

// edit application does not include status because status is edited separately
type EditApplicationRequest struct {
	ID             string `json:"id"`
	Role           string `json:"role"`
	Company        string `json:"company"`
	Location       string `json:"location"`
	AppliedDate    int64  `json:"appliedDate"`
	Link           string `json:"link"`
	OldRole        string `json:"oldRole"`
	OldCompany     string `json:"oldCompany"`
	OldLocation    string `json:"oldLocation"`
	OldLink        string `json:"oldLink"`
	OldAppliedDate int64  `json:"oldAppliedDate"`
	Status         string `json:status`
}

type Handler struct {
	FirestoreClient *firestore.Client
	algoliaClient   *search.APIClient
	pubsubTopic     *pubsub.Topic
	orderingKey     string
}

func NewHandler(
	firestoreClient *firestore.Client,
	algoliaClient *search.APIClient,
	pubsubTopic *pubsub.Topic,
	orderingKey string,
) *Handler {
	return &Handler{
		FirestoreClient: firestoreClient,
		algoliaClient:   algoliaClient,
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
		"timestamp":   time.Now().Add(12 * time.Hour).Truncate(24 * time.Hour).Unix(),
		"objectID":    doc.ID,
	}

	err = h.publishMessage(message)
	if err != nil {
		// delete added application if publish fails
		_, err = h.FirestoreClient.Collection("users").Doc(email).Collection("applications").Doc(doc.ID).Delete(r.Context())
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
	_, err = userDoc.Update(context.Background(), []firestore.Update{
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
	_, err = h.FirestoreClient.Collection("users").Doc(email).Collection("applications").Doc(applicationID).Delete(context.Background())
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
		_, err = h.FirestoreClient.Collection("users").Doc(email).Collection("applications").Doc(applicationID).Set(context.Background(), map[string]interface{}{
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
	_, err = userDoc.Update(context.Background(), []firestore.Update{
		{Path: "applicationsCount", Value: firestore.Increment(-1)},
		{Path: fmt.Sprintf("%s_count", strings.ToLower(deleteApplicationRequest.Status)), Value: firestore.Increment(-1)},
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
		"operation":   "edit",
		"email":       email,
		"objectID":    applicationID,
		"status":      newStatus,
		"appliedDate": appliedDate,
		// since appliedDate is always using noon as the time, we need to ensure
		// that the timestamp sent to PubSub is always at or after noon. this is because
		// a user can edit status of an application at 11:59 AM and the appliedDate is 12:00 PM
		// so this will cause response time metrics to be incorrect
		// so, simply add 12 hours to guarantee it's always at or after noon
		// truncate is used to round down to nearest day just in case the +12 overshoots
		"timestamp": time.Now().Add(12 * time.Hour).Truncate(24 * time.Hour).Unix(),
	}

	err = h.publishMessage(message)
	if err != nil {
		// revert status if publish fails
		_, err = h.FirestoreClient.Collection("users").Doc(email).Collection("applications").Doc(applicationID).Update(context.Background(), []firestore.Update{
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
	_, err = userDoc.Update(context.Background(), []firestore.Update{
		{Path: fmt.Sprintf("%s_count", strings.ToLower(newStatus)), Value: firestore.Increment(1)},
		{Path: fmt.Sprintf("%s_count", strings.ToLower(EditApplicationStatusRequest.OldStatus)), Value: firestore.Increment(-1)},
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

	// just check if we need to update at all use a big ugly switch statement
	// no loops or function calls to reduce memory overhead
	switch {
	case editApplicationRequest.Role != editApplicationRequest.OldRole:
		break
	case editApplicationRequest.Company != editApplicationRequest.OldCompany:
		break
	case editApplicationRequest.Location != editApplicationRequest.OldLocation:
		break
	case editApplicationRequest.Link != editApplicationRequest.OldLink:
		break
	case editApplicationRequest.AppliedDate != editApplicationRequest.OldAppliedDate:
		break
	default:
		log.Println("No application change, returning success")
		w.WriteHeader(http.StatusOK)
		return
	}

	// frontend will send all fields, so we need to update all fields
	_, err = h.FirestoreClient.Collection("users").Doc(email).Collection("applications").Doc(applicationID).Update(r.Context(), []firestore.Update{
		{Path: "role", Value: editApplicationRequest.Role},
		{Path: "company", Value: editApplicationRequest.Company},
		{Path: "location", Value: editApplicationRequest.Location},
		{Path: "link", Value: editApplicationRequest.Link},
		{Path: "appliedDate", Value: editApplicationRequest.AppliedDate},
	})
	if err != nil {
		fmt.Printf("Error editing application: %v\n", err)
		http.Error(w, "Error editing application", http.StatusInternalServerError)
		return
	}

	log.Println("Application edited")

	message := map[string]interface{}{
		"operation":   "edit",
		"email":       email,
		"appliedDate": editApplicationRequest.AppliedDate,
		"company":     editApplicationRequest.Company,
		"link":        editApplicationRequest.Link,
		"location":    editApplicationRequest.Location,
		"role":        editApplicationRequest.Role,
		"status":      editApplicationRequest.Status, // status is only sent to satisfy bigquery schema
		"objectID":    applicationID,
		"timestamp":   time.Now().Add(12 * time.Hour).Truncate(24 * time.Hour).Unix(),
	}

	err = h.publishMessage(message)
	if err != nil {
		// revert status if publish fails
		_, err = h.FirestoreClient.Collection("users").Doc(email).Collection("applications").Doc(applicationID).Update(context.Background(), []firestore.Update{
			{Path: "role", Value: editApplicationRequest.OldRole},
			{Path: "company", Value: editApplicationRequest.OldCompany},
			{Path: "location", Value: editApplicationRequest.OldLocation},
			{Path: "link", Value: editApplicationRequest.OldLink},
			{Path: "appliedDate", Value: editApplicationRequest.OldAppliedDate},
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

// publishes to both algolia and bigquery topics
func (h *Handler) publishMessage(message map[string]interface{}) error {
	messageBody, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("Error marshaling message: %v\n", err)
		return err
	}

	// hold the publish result. this is necessary for
	// our strong consistency model
	var result *pubsub.PublishResult

	// attempt to publish message (algolia and bigquery both subscribe to this topic)
	r := h.pubsubTopic.Publish(context.Background(), &pubsub.Message{
		Data:        messageBody,
		OrderingKey: h.orderingKey,
	})

	// if message publish fails, propagate an error to revert Firestore operation
	result = r
	id, err := result.Get(context.Background())
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
