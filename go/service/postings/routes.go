package postings

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// refer to https://api.cvrve.me/internships/playground
type CvrveAPIResponse struct {
	ID string `json:"id"`
	Company string `json:"company_name"`
	Title string `json:"title"`
	Locations []string `json:"locations"`
	PostedDate int64 `json:"date_posted"`
	UpdatedDate int64 `json:"date_updated"`
	Active bool `json:"active"`
	IsVisible bool `json:"is_visible"`
	Sponsorship string `json:"sponsorship"`
	Url string `json:"url"`
}

// need total pages and current pages for nice looking pagination
// vansh i already implemented keyset + offset hybrid pagination
// let me integrate bro :handshake:
type PostingsResponse struct {
	Applications []CvrveAPIResponse `json:"postings"`
	TotalPages   int               	`json:"totalPages"`
	CurrentPage  int               	`json:"currentPage"`
}

type Handler struct {
	// nothinggg unless in the future we need to know if user has premium or whatevuh
}

func NewHandler() *Handler {
	return &Handler{
		// this has nothing, but just for consistency we keep the handler if later on
		// some dependency is needed in postings. otherwise its just simple api call
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/postings", h.GetPostings).Methods("GET")
}

func (h *Handler) GetPostings(w http.ResponseWriter, r *http.Request) {
	log.Println("[*] GetPostings [*]")
	log.Println("-----------------")

	// user does not have to be authed to access this

	// fetch from cvrve api with query params. api key is in env
	// params: intern || new grad, company, title, location, sponsorship, active, page, hitsPerPage 

	log.Println("(not implemented): fetch from cvrve api instead")
	log.Println("-----------------")

	// build response object 
	postingsResponse := PostingsResponse{
		Applications: nil,
		TotalPages:   1,
		CurrentPage:  1,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(postingsResponse)
}
