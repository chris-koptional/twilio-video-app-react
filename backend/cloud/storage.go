package cloud

import (
	"context"

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
