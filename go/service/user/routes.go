package user

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/juhun32/jtracker-backend/service/auth"
	"github.com/juhun32/jtracker-backend/utils"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/api/iterator"
)

type AddApplicationRequest struct {
	Role        string `json:"role"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	AppliedDate string `json:"appliedDate"`
	Status      string `json:"status"`
	Link        string `json:"link"`
}

type Application struct {
	ID          string `json:"id"`
	Role        string `json:"role"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	AppliedDate string `json:"appliedDate"`
	Status      string `json:"status"`
	Link        string `json:"link"`
}

type DeleteApplicationRequest struct {
	ID string `json:"id"`
}

type EditStatusApplicationRequest struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// edit application does not include status because status is edited separately
type EditApplicationRequest struct {
	ID          string `json:"id"`
	Role        string `json:"role"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	AppliedDate string `json:"appliedDate"`
	Link        string `json:"link"`
}

type Handler struct {
	firestoreClient *firestore.Client
	rabbitCh        *amqp.Channel
	rabbitQ         amqp.Queue
}

func NewHandler(firestoreClient *firestore.Client,
	rabbitCh *amqp.Channel,
	rabbitQ amqp.Queue,
) *Handler {
	return &Handler{
		firestoreClient: firestoreClient,
		rabbitCh:        rabbitCh,
		rabbitQ:         rabbitQ,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/user/dashboard", h.Dashboard).Methods("GET").Name("dashboard")
	router.HandleFunc("/user/addApplication", h.AddApplication).Methods("POST").Name("addApplication")
	router.HandleFunc("/user/deleteApplication", h.DeleteApplication).Methods("POST").Name("deleteApplication")
	router.HandleFunc("/user/editStatus", h.EditStatus).Methods("POST").Name("editStatus")
	router.HandleFunc("/user/editApplication", h.EditApplication).Methods("POST").Name("editApplication")
}

func (h *Handler) AddApplication(w http.ResponseWriter, r *http.Request) {
	user, err := auth.IsAuthenticated(r)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// for user's email (unique ID), add application and assign unique jobID
	email := user.Email

	fmt.Println(email)

	// extract json from request body
	var addApplicationRequest AddApplicationRequest
	err = json.NewDecoder(r.Body).Decode(&addApplicationRequest)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	// add application to Firestore using users/{email} where jobs is a document within the user's collection
	doc, _, err := h.firestoreClient.Collection("users").Doc(email).Collection("applications").Add(r.Context(), map[string]interface{}{
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
	w.WriteHeader(http.StatusCreated)

	// for now, just send msg to rabbitmq (later, we will ensure that the application is indexed)
	message := map[string]string{
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

	messageBody, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("Error marshaling message: %v\n", err)
		return
	}

	err = h.rabbitCh.Publish(
		"",             // exchange
		h.rabbitQ.Name, // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        messageBody,
		})
	if err != nil {
		fmt.Printf("Error publishing message: %v\n", err)
	}
}

// current implementation is TEMPORARY!!!!
// actual implementation doesn't query Firestore, it queries Algolia
// Firestore is just a backup in case we want to switch to a diff search engine
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	user, err := auth.IsAuthenticated(r)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fmt.Println("user", user)

	// the actual implementation of this will use the user object from auth.IsAuthenticated
	// TEMP: query firestore for user's applications
	// TODO: query Algolia for user's applications instead
	email := user.Email

	iter := h.firestoreClient.Collection("users").Doc(email).Collection("applications").Documents(r.Context())
	var applications []Application

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			http.Error(w, "Error querying Firestore", http.StatusInternalServerError)
			return
		}

		application := doc.Data()
		applications = append(applications, Application{
			ID:          doc.Ref.ID,
			Role:        utils.GetString(application, "role"),
			Company:     utils.GetString(application, "company"),
			Location:    utils.GetString(application, "location"),
			AppliedDate: utils.GetString(application, "appliedDate"),
			Status:      utils.GetString(application, "status"),
			Link:        utils.GetString(application, "link"), // link might be nil so getString
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(applications)
}

func (h *Handler) DeleteApplication(w http.ResponseWriter, r *http.Request) {
	user, err := auth.IsAuthenticated(r)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	email := user.Email

	// extract json from request body
	var deleteApplicationRequest DeleteApplicationRequest
	err = json.NewDecoder(r.Body).Decode(&deleteApplicationRequest)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	applicationID := deleteApplicationRequest.ID

	// delete application from Firestore
	_, err = h.firestoreClient.Collection("users").Doc(email).Collection("applications").Doc(applicationID).Delete(r.Context())
	if err != nil {
		fmt.Printf("Error deleting application: %v\n", err)
		http.Error(w, "Error deleting application", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	// for now, just send msg to rabbitmq (later, we will ensure that the application is indexed)
	err = h.rabbitCh.Publish(
		"",             // exchange
		h.rabbitQ.Name, // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte("Application deleted"),
		})
	if err != nil {
		fmt.Printf("Error publishing message: %v\n", err)
	}
}

func (h *Handler) EditStatus(w http.ResponseWriter, r *http.Request) {
	user, err := auth.IsAuthenticated(r)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	email := user.Email

	// extract json from request body
	var editStatusApplicationRequest EditStatusApplicationRequest
	err = json.NewDecoder(r.Body).Decode(&editStatusApplicationRequest)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	applicationID := editStatusApplicationRequest.ID

	// edit application in Firestore
	_, err = h.firestoreClient.Collection("users").Doc(email).Collection("applications").Doc(applicationID).Update(r.Context(), []firestore.Update{
		{
			Path:  "status",
			Value: editStatusApplicationRequest.Status,
		},
	})
	if err != nil {
		fmt.Printf("Error adding application: %v\n", err)
		http.Error(w, "Error adding application", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	// for now, just send msg to rabbitmq (later, we will ensure that the application is indexed)
	err = h.rabbitCh.Publish(
		"",             // exchange
		h.rabbitQ.Name, // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte("Application status edited"),
		})
	if err != nil {
		fmt.Printf("Error publishing message: %v\n", err)
	}
}

func (h *Handler) EditApplication(w http.ResponseWriter, r *http.Request) {
	user, err := auth.IsAuthenticated(r)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	email := user.Email

	// extract json from request body
	var addApplicationRequest EditApplicationRequest
	err = json.NewDecoder(r.Body).Decode(&addApplicationRequest)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	applicationID := addApplicationRequest.ID

	// frontend will send all fields, so we need to update all fields
	_, err = h.firestoreClient.Collection("users").Doc(email).Collection("applications").Doc(applicationID).Update(r.Context(), []firestore.Update{
		{
			Path:  "role",
			Value: addApplicationRequest.Role,
		},
		{
			Path:  "company",
			Value: addApplicationRequest.Company,
		},
		{
			Path:  "location",
			Value: addApplicationRequest.Location,
		},
		{
			Path:  "link",
			Value: addApplicationRequest.Link,
		},
		{
			Path:  "appliedDate",
			Value: addApplicationRequest.AppliedDate,
		},
	})
	if err != nil {
		fmt.Printf("Error editing application: %v\n", err)
		http.Error(w, "Error editing application", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	// for now, just send msg to rabbitmq (later, we will ensure that the application is indexed)
	err = h.rabbitCh.Publish(
		"",             // exchange
		h.rabbitQ.Name, // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte("Application edited" + applicationID + " " + email),
		})
	if err != nil {
		fmt.Printf("Error publishing message: %v\n", err)
	}
}
