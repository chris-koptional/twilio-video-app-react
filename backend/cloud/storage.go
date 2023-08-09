package cloud

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

func CreateStorageClient() (*storage.Client, error) {

	credentials := create_secret_JSON()

	opts := option.WithCredentialsJSON(credentials)

	return storage.NewClient(context.Background(), opts)
}

func GetTranscriptionObject(c *storage.Client, documentId string) (string, error) {

	bucket := os.Getenv("TRANSCRIPTION_BUCKET")

	r, err := c.Bucket(bucket).Object(documentId + ".txt").NewReader(context.Background())

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
