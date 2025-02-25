package postings

import (
	"log"
	"fmt"
	"net/http"
	"github.com/juhun32/jtracker-backend/service/auth"

	"github.com/gorilla/mux"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
)

type Handler struct {
	algoliaClient *search.APIClient
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
	// 1.b) get the page number from request
	
	// 2. build a search params object (see users/dashboard as a reference)
	// 		(IM PRETTY SURE) we don't need Facets and FacetFilters param 
	// 2.a) set free text query if provided
	// 2.b) finalize search params object

	// 3. query algolia w/ search params

	// 4.a) extract hits
	// 4.b) marshal raw hits into json
	// 4.c) unmarshal json into algoliaresponse slice 

	// 5. return 
}

