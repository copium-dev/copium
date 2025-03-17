package prod

// the production version; aka push-based subscription instead of pull-based
import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "sync/atomic"

    "github.com/copium-dev/copium/algolia-consumer/inits"
    "github.com/copium-dev/copium/algolia-consumer/job"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
)

var (
	algoliaClient *search.APIClient
	counter int32 = 1
)

type PubSubMessage struct {
    Message struct {
        Data []byte `json:"data,omitempty"`
        ID   string `json:"id"`
        Attributes map[string]string `json:"attributes,omitempty"`
    } `json:"message"`
    Subscription string `json:"subscription"`
}

func main() {
	var err error 

	// initialize algolia client
	algoliaClient, err = inits.InitializeAlgoliaClient()
	if err != nil {
		log.Fatalf("Error initializing Algolia client: %v", err)
	}

	// push-based subscription, so set up HTTP server
	http.HandleFunc("/", pubsubHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("[*] ALGOLIA [*] Starting push subscription server on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

func pubsubHandler(w http.ResponseWriter, r *http.Request) {
	// only allow post requests
	if r.Method != http.MethodPost {
		http.Error(w, "405 - Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// parse pubsub message using PubSubMessage struct
	var pubsubMessage PubSubMessage
	if err := json.NewDecoder(r.Body).Decode(&pubsubMessage); err != nil {
		log.Printf("[*] ALGOLIA [*] Error decoding Pub/Sub message: %v", err)
		http.Error(w, "400 - Bad Request", http.StatusBadRequest)
		return
	}

	log.Printf("[*] ALGOLIA [*] Received Pub/Sub message: %s", pubsubMessage.Message.Data)

	// do the processing as normal (the same logic as the callback in pull-based subscription)
	jobID := atomic.AddInt32(&counter, 1)
	newJob, err := job.NewJob(pubsubMessage.Message.Data, jobID, algoliaClient)
	if err != nil {
		log.Printf("Failed to create job %d: %s", jobID, err)
		// return 4xx for non-retryable and 5xx for retryable
        http.Error(w, fmt.Sprintf("Failed to create job: %v", err), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	err = newJob.Process(ctx)
	if err != nil {
		log.Printf("Failed to process job %d: %s", jobID, err)
		// return 4xx for non-retryable and 5xx for retryable
		http.Error(w, fmt.Sprintf("Failed to process job: %v", err), http.StatusInternalServerError)
	}

	fmt.Println("Job done, acking message (ALGOLIA)")
	// pubsub push-based ACK is just returning 200 OK
	w.WriteHeader(http.StatusOK)
}