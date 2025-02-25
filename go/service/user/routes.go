package user

// this file contains the HTTP handlers for the user service
// it contains the following handlers:
// (R) - Dashboard: queries Algolia for applications based on search query
// (R) - Profile: (for now) returns simply email and app count; once we figure out what kind of data analytics we want to show, it will be updated
// (C) - AddApplication: adds an application to Firestore and publishes a message to RabbitMQ
// (D) - DeleteApplication: deletes an application from Firestore and publishes a message to RabbitMQ
// (U) - EditStatus: edits the status of an application in Firestore and publishes a message to RabbitMQ
// (U) - EditApplication: edits an application in Firestore and publishes a message to RabbitMQ
// (D) - DeleteUser: deletes a user from Firestore and publishes a message to RabbitMQ to delete all applications from Algolia
// this file contains the following utility functions:
// - deleteUserFromFirestore: deletes a user from Firestore, including all applications
// - publishMessage: publishes a message to RabbitMQ with publish and connection retries
//     (relies on utils.PublishWithRetry and retryRabbitConnectionAndRetryPublish)
// - retryRabbitConnectionAndRetryPublish: re-establishes connection to RabbitMQ and retries publishing a message
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

	"github.com/juhun32/jtracker-backend/service/auth"
	"github.com/juhun32/jtracker-backend/utils"
	"github.com/juhun32/jtracker-backend/service/user/userutils"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
	"cloud.google.com/go/pubsub"
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
	Applications []AlgoliaResponse 	`json:"applications"`
	TotalPages   int           		`json:"totalPages"`
	CurrentPage  int           		`json:"currentPage"`
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
	ID string `json:"id"`
	Role string `json:"role"`
	Company string `json:"company"`
	Location string `json:"location"`
	AppliedDate int64 `json:"appliedDate"`
	Status string `json:"status"`
	Link string `json:"link"`
}

type EditApplicationStatusRequest struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	OldStatus string `json:"oldStatus"`
}

// edit application does not include status because status is edited separately
type EditApplicationRequest struct {
	ID          string `json:"id"`
	Role        string `json:"role"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	AppliedDate int64  `json:"appliedDate"`
	Link        string `json:"link"`
	OldRole     string `json:"oldRole"`
	OldCompany  string `json:"oldCompany"`
	OldLocation string `json:"oldLocation"`
	OldLink     string `json:"oldLink"`
	OldAppliedDate int64 `json:"oldAppliedDate"`
}

type Handler struct {
	FirestoreClient *firestore.Client
	algoliaClient   *search.APIClient
	pubsubClient   *pubsub.Client
}

func NewHandler(
	firestoreClient *firestore.Client,
	algoliaClient *search.APIClient,
	pubsubClient *pubsub.Client,
) *Handler {
	return &Handler{
		FirestoreClient: firestoreClient,
		algoliaClient:   algoliaClient,
		pubsubClient:   pubsubClient,
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

	user, err := auth.IsAuthenticated(r)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Println("User authenticated")

	// get user's email
	email := user.Email

	// get user's applications count
	doc, err := h.FirestoreClient.Collection("users").Doc(email).Get(r.Context())
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Error retrieving user data", http.StatusInternalServerError)
		return
	}
	applicationsCount := doc.Data()["applicationsCount"].(int64)
	log.Println("Applications count retrieved")
	log.Println("-----------------")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"email":             email,
		"applicationsCount": applicationsCount,
	})
}

// queries algolia for applications based on search query
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	log.Println("[*] Dashboard [*]")
	log.Println("-----------------")

	user, err := auth.IsAuthenticated(r)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Println("User authenticated")

	email := user.Email
	fmt.Println(email)

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

	// 2. build a search params object
	searchParamsObject := &search.SearchParamsObject{
		Facets:       []string{"email"},
		FacetFilters: &search.FacetFilters{String: utils.StringPtr("email:" + email)},
		HitsPerPage:  utils.IntPtr(10),
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
		TotalPages:   userutils.CalculateTotalPages(int(*response.NbHits), 10),
		CurrentPage:  page,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseObject)
}

// consistency: if publish fails, delete the application from Firestore
func (h *Handler) AddApplication(w http.ResponseWriter, r *http.Request) {
	log.Println("[*] AddApplication [*]")
	log.Println("-----------------")

	user, err := auth.IsAuthenticated(r)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// for user's email (unique ID), add application and assign unique jobID
	email := user.Email

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

	w.WriteHeader(http.StatusCreated)
	if f, ok := w.(http.Flusher); ok {
		f.Flush() // flushes the headers immediately (below are background operations)
	}

	message := map[string]interface{}{
		"operation":   "add",
		"email":       email,
		"appliedDate": addApplicationRequest.AppliedDate,
		"company":     addApplicationRequest.Company,
		"link":        addApplicationRequest.Link,
		"location":    addApplicationRequest.Location,
		"role":        addApplicationRequest.Role,
		"status":      addApplicationRequest.Status,
		"objectID":    doc.ID,
	}

	err = h.publishMessage(message); if err != nil {
		// delete added application if publish fails
		_, err = h.FirestoreClient.Collection("users").Doc(email).Collection("applications").Doc(doc.ID).Delete(r.Context())
		if err != nil {
			fmt.Printf("Error deleting application: %v\n", err)
		}
		log.Println("AddApplication reverted because of publish failure")
		return
	}

	// to reduce amount of reads in this single request, increment applicationCount AFTER verifying publish success
	// this means we don't have to revert applicationCount on top of reverting the add operation. this DOES
	// introduce small window of inconsistency but this is reducing costs and reducing complexity
	userDoc := h.FirestoreClient.Collection("users").Doc(email)
	_, err = userDoc.Update(context.Background(), []firestore.Update{
		{Path: "applicationsCount", Value: firestore.Increment(1)},
	})
	// can't return HTTP error if we already wrote StatusCreated...
	if err != nil {
		fmt.Printf("Error updating applications count: %v\n", err)
		return
	}

	log.Println("Applications count updated, added by 1")
}

// consistency: save current application status; if publish fails, revert status
func (h *Handler) DeleteApplication(w http.ResponseWriter, r *http.Request) {
	log.Println("[*] DeleteApplication [*]")
	log.Println("-----------------")

	user, err := auth.IsAuthenticated(r)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	email := user.Email

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

	w.WriteHeader(http.StatusOK)
	if f, ok := w.(http.Flusher); ok {
		f.Flush() // flushes the headers immediately (below are background operations)
	}

	message := map[string]interface{}{
		"operation": "delete",
		"email":     email,
		"objectID":  applicationID,
	}

	err = h.publishMessage(message); if err != nil {
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
		}
		log.Println("DeleteApplication reverted because of publish failure")
		return
	}

	// to reduce amount of reads in this single request, decrement applicationCount AFTER verifying publish success
	// this means we don't have to revert applicationCount on top of reverting the delete operation. this DOES
	// introduce small window of inconsistency but this is reducing costs and reducing complexity
	userDoc := h.FirestoreClient.Collection("users").Doc(email)
	_, err = userDoc.Update(context.Background(), []firestore.Update{
		{Path: "applicationsCount", Value: firestore.Increment(-1)},
	})
	// can't return HTTP error if we already wrote StatusOK...
	if err != nil {
		fmt.Printf("Error updating applications count: %v\n", err)
		return
	}
}

func (h *Handler) EditStatus(w http.ResponseWriter, r *http.Request) {
	log.Println("[*] EditStatus [*]")
	log.Println("-----------------")

	user, err := auth.IsAuthenticated(r)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	email := user.Email

	log.Println("User authenticated")

	// extract json from request body
	var EditApplicationStatusRequest EditApplicationStatusRequest
	err = json.NewDecoder(r.Body).Decode(&EditApplicationStatusRequest)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	applicationID := EditApplicationStatusRequest.ID

	_, err = h.FirestoreClient.Collection("users").Doc(email).Collection("applications").Doc(applicationID).Update(r.Context(), []firestore.Update{
		{
			Path:  "status",
			Value: EditApplicationStatusRequest.Status,
		},
	})
	if err != nil {
		fmt.Printf("Error adding application: %v\n", err)
		http.Error(w, "Error adding application", http.StatusInternalServerError)
		return
	}

	log.Println("Application status edited")

	w.WriteHeader(http.StatusCreated)
	if f, ok := w.(http.Flusher); ok {
		f.Flush() // flushes the headers immediately (below are background operations)
	}

	message := map[string]interface{}{
		"operation": "edit",
		"email":     email,
		"objectID":  applicationID,
		"status":    EditApplicationStatusRequest.Status,
	}

	err = h.publishMessage(message); if err != nil {
		// revert status if publish fails 
		_, err = h.FirestoreClient.Collection("users").Doc(email).Collection("applications").Doc(applicationID).Update(context.Background(), []firestore.Update{
			{
				Path: "status",
				Value: EditApplicationStatusRequest.OldStatus,
			},
		})
		if err != nil {
			fmt.Printf("Error reverting status: %v\n", err)
		}
		log.Println("EditStatus reverted because of publish failure")
		return
	}
}

func (h *Handler) EditApplication(w http.ResponseWriter, r *http.Request) {
	log.Println("[*] EditApplication [*]")
	log.Println("-----------------")

	user, err := auth.IsAuthenticated(r)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	email := user.Email

	log.Println("User authenticated")

	// extract json from request body
	var addApplicationRequest EditApplicationRequest
	err = json.NewDecoder(r.Body).Decode(&addApplicationRequest)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	applicationID := addApplicationRequest.ID

	// frontend will send all fields, so we need to update all fields
	_, err = h.FirestoreClient.Collection("users").Doc(email).Collection("applications").Doc(applicationID).Update(r.Context(), []firestore.Update{
		{Path:  "role", Value: addApplicationRequest.Role},
		{Path:  "company", Value: addApplicationRequest.Company},
		{Path:  "location", Value: addApplicationRequest.Location},
		{Path:  "link", Value: addApplicationRequest.Link},
		{Path:  "appliedDate", Value: addApplicationRequest.AppliedDate},
	})
	if err != nil {
		fmt.Printf("Error editing application: %v\n", err)
		http.Error(w, "Error editing application", http.StatusInternalServerError)
		return
	}

	log.Println("Application edited")

	w.WriteHeader(http.StatusOK)
	if f, ok := w.(http.Flusher); ok {
		f.Flush() // flushes the headers immediately (below are background operations)
	}

	message := map[string]interface{}{
		"operation":   "edit",
		"email":       email,
		"appliedDate": addApplicationRequest.AppliedDate,
		"company":     addApplicationRequest.Company,
		"link":        addApplicationRequest.Link,
		"location":    addApplicationRequest.Location,
		"role":        addApplicationRequest.Role,
		"objectID":    applicationID,
	}

	err = h.publishMessage(message); if err != nil {
		// revert status if publish fails
		_, err = h.FirestoreClient.Collection("users").Doc(email).Collection("applications").Doc(applicationID).Update(context.Background(), []firestore.Update{
			{Path: "role", Value: addApplicationRequest.OldRole},
			{Path: "company", Value: addApplicationRequest.OldCompany},
			{Path: "location", Value: addApplicationRequest.OldLocation},
			{Path: "link", Value: addApplicationRequest.OldLink},
			{Path: "appliedDate", Value: addApplicationRequest.OldAppliedDate},
		})
		if err != nil {
			fmt.Printf("Error reverting application: %v\n", err)
		}
		log.Println("EditApplication reverted because of publish failure")
		return
	}
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	log.Println("[*] DeleteUser [*]")
	log.Println("-----------------")

	user, err := auth.IsAuthenticated(r)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	email := user.Email

	log.Println("User authenticated")

	// send to algolia to delete all applications associated with this user
	message := map[string]interface{}{
		"operation": "userDelete",
		"email":     email,
	}

	err = h.publishMessage(message); if err != nil {
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
	r := h.pubsubClient.Topic("applications").Publish(context.Background(), &pubsub.Message{
		Data: messageBody,
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