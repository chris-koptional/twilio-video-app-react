package cloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	taskspb "cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
	"google.golang.org/api/option"
)

type ServiceAccountCredentials struct {
	Type           string `json:"type"`
	ProjectID      string `json:"project_id"`
	PrivateKeyID   string `json:"private_key_id"`
	PrivateKey     string `json:"private_key"`
	ClientEmail    string `json:"client_email"`
	ClientID       string `json:"client_id"`
	AuthURI        string `json:"auth_uri"`
	TokenURI       string `json:"token_uri"`
	AuthCert       string `json:"auth_provider_x509_cert_url"`
	ClientCert     string `json:"client_x509_cert_url"`
	UniverseDomain string `json:"universe_domain"`
}

type TaskOptions struct {
	Queue       QueueDetails
	Payload     any
	CallbackURL string
}

type QueueDetails struct {
	ProjectID string
	Location  string
	Name      string
}

func CreatePostRecordingTask(client *cloudtasks.Client, callbackURL string, payload any) (*taskspb.Task, error) {
	queue := QueueDetails{
		ProjectID: os.Getenv("PROJECT_ID"),
		Location:  "us-east4",
		Name:      "post-recording",
	}
	task := TaskOptions{
		CallbackURL: callbackURL,
		Payload:     payload,
		Queue:       queue,
	}

	return createHTTPTask(client, task)
}

func createHTTPTask(client *cloudtasks.Client, task TaskOptions) (*taskspb.Task, error) {
	ctx := context.Background()
	// Build the Task queue path.
	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", task.Queue.ProjectID, task.Queue.Location, task.Queue.Name)

	// Build the Task payload.
	req := &taskspb.CreateTaskRequest{
		Parent: queuePath,
		Task: &taskspb.Task{
			MessageType: &taskspb.Task_HttpRequest{
				HttpRequest: &taskspb.HttpRequest{
					HttpMethod: taskspb.HttpMethod_POST,
					Url:        task.CallbackURL,
				},
			},
		},
	}

	// Add a payload message if one is present.
	jsonPayload, err := json.Marshal(task.Payload)

	if err != nil {
		return nil, errors.New("Could not marshal payload")
	}
	req.Task.GetHttpRequest().Body = jsonPayload

	createdTask, err := client.CreateTask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("cloudtasks.CreateTask: %w", err)
	}

	return createdTask, nil
}

func create_secret_JSON() []byte {

	credentialType := os.Getenv("CREDENTIAL_TYPE")
	projectId := os.Getenv("PROJECT_ID")
	privateKeyID := os.Getenv("PRIVATE_KEY_ID")
	privateKey := strings.ReplaceAll(os.Getenv("PRIVATE_KEY"), "\\n", "\n")
	clientEmail := os.Getenv("CLIENT_EMAIL")
	clientID := os.Getenv("CLIENT_ID")
	authURI := os.Getenv("AUTH_URI")
	tokenURI := os.Getenv("TOKEN_URI")
	authProvider := os.Getenv("AUTH_PROVIDER_X509_CERT_URL")
	clientCert := os.Getenv("CLIENT_X509_CERT_URL")
	universeDomain := os.Getenv("UNIVERSE_DOMAIN")

	credentials := ServiceAccountCredentials{
		Type:           credentialType,
		ProjectID:      projectId,
		PrivateKeyID:   privateKeyID,
		PrivateKey:     privateKey,
		ClientEmail:    clientEmail,
		ClientID:       clientID,
		AuthURI:        authURI,
		TokenURI:       tokenURI,
		AuthCert:       authProvider,
		ClientCert:     clientCert,
		UniverseDomain: universeDomain,
	}

	jsonBytes, err := json.Marshal(credentials)

	if err != nil {
		log.Fatalf("Failed to stringify secret credentials", err.Error())

	}

	return jsonBytes
}

func CreateClient() (*cloudtasks.Client, error) {
	ctx := context.Background()

	credentials := create_secret_JSON()

	opts := option.WithCredentialsJSON(credentials)
	return cloudtasks.NewClient(ctx, opts)
}
