package postings

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/juhun32/jtracker-backend/service/auth"
	"github.com/juhun32/jtracker-backend/service/postings/postingsutils"
	"github.com/juhun32/jtracker-backend/utils"

	"github.com/gorilla/mux"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
)

type Handler struct {
	algoliaClient *search.APIClient
}

type AlgoliaResponse struct {
	Company     string   `json:"company_name"`
	Locations   []string `json:"locations"`
	Title       string   `json:"title"`
	PostedDate  int64    `json:"date_posted"`
	UpdatedDate int64    `json:"date_updated"`
	Url			string   `json:"url"`
}

type PostingsResponse struct {
	Applications []AlgoliaResponse `json:"postings"`
	TotalPages   int               `json:"totalPages"`
	CurrentPage  int               `json:"currentPage"`
}

func NewHandler(algoliaClient *search.APIClient) *Handler {
	return &Handler{
		algoliaClient: algoliaClient,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/postings", h.GetPostings).Methods("GET")
}

func (h *Handler) GetPostings(w http.ResponseWriter, r *http.Request) {
	log.Println("[*] GetPostings [*]")
	log.Println("-----------------")

	// the ONLY point of auth is so that only logged in users can access this endpoint
	_, err := auth.IsAuthenticated(r)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Println("User authenticated")

	// 1.a) extract search query from request
	// 		we need a different function than userutils.ParseQuery (so maybe make a postingutils package)
	queryText, filtersString, err := postingsutils.ParseQuery(r)

	// 1.b) get the page number from request
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

	// 2. build a search params object (see users/dashboard as a reference)
	// 		(IM PRETTY SURE) we don't need Facets and FacetFilters param
	// 2.a) set free text query if provided
	// 2.b) finalize search params object

	searchParamsObject := &search.SearchParamsObject{
		HitsPerPage: utils.IntPtr(10),
		Filters:     utils.StringPtr(filtersString),
		Page:        utils.IntPtr(int32(page)),
	}

	// set free text query if present
	if queryText != "" {
		searchParamsObject.Query = utils.StringPtr(queryText)
		log.Println("Free text query text extracted: ", queryText)
	}

	searchParams := &search.SearchParams{
		SearchParamsObject: searchParamsObject,
	}

	// 3. query algolia w/ search params
	response, err := h.algoliaClient.SearchSingleIndex(
		h.algoliaClient.NewApiSearchSingleIndexRequest("postings").WithSearchParams(searchParams),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		http.Error(w, "Error querying Algolia", http.StatusInternalServerError)
		return
	}

	// 4.a) extract hits
	var applications []AlgoliaResponse

	// 4.b) marshal raw hits into json
	hitsBytes, err := json.Marshal(response.Hits)
	if err != nil {
		fmt.Printf("Error marshaling hits: %v\n", err)
		http.Error(w, "Error processing hits", http.StatusInternalServerError)
		return
	}

	// 4.c) unmarshal json into algoliaresponse slice
	err = json.Unmarshal(hitsBytes, &applications)
	if err != nil {
		fmt.Printf("Error unmarshaling hits: %v\n", err)
		http.Error(w, "Error processing applications", http.StatusInternalServerError)
		return
	}

	log.Println("Postings extracted:", applications)
	log.Println("-----------------")

	// 5. return
	responseObject := PostingsResponse{
		Applications: applications,
		TotalPages:   postingsutils.CalculateTotalPages(int(*response.NbHits), 10),
		CurrentPage:  page,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseObject)
}
