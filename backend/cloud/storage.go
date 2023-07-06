package cloud

import (
	"context"
	"errors"

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
