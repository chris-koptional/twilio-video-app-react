package cloud

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

func CreateStorageClient() (*storage.Client, error) {

	credentials := create_secret_JSON()

	opts := option.WithCredentialsJSON(credentials)

	return storage.NewClient(context.Background(), opts)
}

func InjectStorageClient(client *storage.Client) gin.HandlerFunc {

	return func(c *gin.Context) {
		c.Set("storage_client", client)
		c.Next()
	}
}

func GetStorageClient(c *gin.Context) (*storage.Client, error) {
	client, ok := c.Get("storage_client")

	if !ok {
		return nil, errors.New("Failed to get storage client")
	}

	taskClient, ok := client.(*storage.Client)

	if !ok {
		return nil, errors.New("Storage client incorrect type")
	}

	return taskClient, nil
}

func GetTranscriptionObject(c *gin.Context, documentId string) (string, error) {

	client, err := GetStorageClient(c)

	if err != nil {
		return "", err
	}

	bucket := os.Getenv("TRANSCRIPTION_BUCKET")

	r, err := client.Bucket(bucket).Object(documentId + ".txt").NewReader(c)

	if err != nil {
		fmt.Println("Failed to create reader.")
		return "", err
	}
	defer r.Close()

	transcript, err := ioutil.ReadAll(r)
	if err != nil {
		fmt.Println("Failed to read file contents.")
		return "", err
	}

	return string(transcript), nil
}
