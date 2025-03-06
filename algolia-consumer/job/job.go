package job

import (
	"encoding/json"
	"fmt"
	"log"
	"context"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
)

type Job struct {
    ID              int32
    Data            map[string]interface{}
    RawData         []byte
    Operation       string
	AlgoliaClient  *search.APIClient
}

// all this really does is unmarshal the raw data and figure out the operation
func NewJob(data []byte, id int32, algoliaClient *search.APIClient) (*Job, error) {
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
		AlgoliaClient:   algoliaClient,
    }, nil
}

func (j *Job) Process(ctx context.Context) error {
	// context is not required for any operations but this is just a security check
	// to make sure we dont process jobs after their context has been cancelled
	if ctx.Err() != nil {
		return ctx.Err()
	}

    log.Printf("[*] Algolia [*]")
    log.Printf("-------")
    log.Printf("Processing Job With ID [%d] with content: [%s]", j.ID, j.Data)
    
    switch j.Operation {
	case "add":
		return j.addApplication(ctx)
	case "edit":
		return j.editApplication(ctx)
	case "delete":
		return j.deleteApplication(ctx)
	case "userDelete":
		return j.userDelete(ctx)
    default:
        return fmt.Errorf("unknown operation: %s", j.Operation)
    }
}

func (j *Job) addApplication(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	data := j.Data

	// add the application to algolia
	saveRes, err := j.AlgoliaClient.SaveObject(
		j.AlgoliaClient.NewApiSaveObjectRequest("users", data),
	)
	if err != nil {
		log.Printf("Failed to save object: %s", err)
		return err
	}

	// wait for task to finish before exiting function
	_, err = j.AlgoliaClient.WaitForTask("users", saveRes.TaskID)
	if err != nil {
		log.Printf("Error waiting for task to finish: %s", err)
		return err
	}

	log.Printf("Saved object: %v", saveRes)
	return nil
}

func (j *Job) editApplication(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	data := j.Data

	// edit the application in algolia
	objectID, ok := data["objectID"].(string)
	if !ok {
		return fmt.Errorf("failed to get objectID from data")
	}

	updateRes, err := j.AlgoliaClient.PartialUpdateObject(
		j.AlgoliaClient.NewApiPartialUpdateObjectRequest("users", objectID, data),
	)
	if err != nil {
		log.Printf("Failed to update object: %s", err)
		return err
	}

	// wait for task to finish before exiting function
	_, err = j.AlgoliaClient.WaitForTask("users", *updateRes.TaskID)
	if err != nil {
		log.Printf("Error waiting for task to finish: %s", err)
		return err
	}

	log.Printf("Updated object: %v", updateRes)
	return nil
}

func (j *Job) deleteApplication(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	data := j.Data

	// delete the application from algolia
	objectID, ok := data["objectID"].(string)
	if !ok {
		return fmt.Errorf("failed to get objectID from data")
	}

	deleteRes, err := j.AlgoliaClient.DeleteObject(
		j.AlgoliaClient.NewApiDeleteObjectRequest("users", objectID),
	)
	if err != nil {
		log.Printf("Failed to delete object: %s", err)
		return err
	}

	// wait for task to finish before exiting function
	_, err = j.AlgoliaClient.WaitForTask("users", deleteRes.TaskID)
	if err != nil {
		log.Printf("Error waiting for task to finish: %s", err)
		return err
	}
	
	log.Printf("Deleted object: %v", deleteRes)
	return nil
}

// note: DeleteBy is resource intensive so we should carefully monitor
func (j *Job) userDelete(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	data := j.Data

	// extract and delete every objectID where email == data["email"]
	email, ok := data["email"].(string)
	if !ok {
		return fmt.Errorf("failed to get email from data")
	}
	
	filter := fmt.Sprintf("email:%s", email)

	res, err := j.AlgoliaClient.DeleteBy(
		j.AlgoliaClient.NewApiDeleteByRequest(
			"users",
			search.NewEmptyDeleteByParams().SetFilters(filter),
		),
	)
	if err != nil {
		log.Printf("Failed to delete by: %s", err)
		return err
	}

	// wait for task to finish before exiting function
	_, err = j.AlgoliaClient.WaitForTask("users", res.TaskID)
	if err != nil {
		log.Printf("Error waiting for task to finish: %s", err)
		return err
	}

	log.Printf("Deleted objects: %s", res)
	return nil
}