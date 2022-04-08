package storage

import (
	"context"
	"net/url"
	"io/ioutil"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"

	"github.com/equinor/oneseismic/api/internal"
)

/*
 *  Minimal interface for fetching blobs/files from storage. This hides a lot of
 *  feature and details about the underlying storage from the rest of the
 *  system, making it easy to swap out the storage provider. This means testing
 *  becomes easier through custom storage implementations.
 */
type StorageClient interface {
	/*
	 * Get a blob or file from storage
	 */
	Get(ctx context.Context, bloburl *url.URL) ([]byte, error)
}

/*
 * Azure Blob Store implementation of a oneseismic StorageClient
 */
type AzStorage struct {
}

func (c *AzStorage) Get(ctx context.Context, bloburl *url.URL) ([]byte, error) {
	if bloburl == nil {
		return []byte{}, internal.InternalError("blob URL is nil")
	}

	client, err := azblob.NewBlobClientWithNoCredential(bloburl.String(), nil)
	if err != nil {
		return []byte{}, internal.InternalError(err.Error())
	}

	dl, err := client.Download(ctx, nil)
	if err != nil {
		return []byte{}, err
	}

	body := dl.Body(&azblob.RetryReaderOptions{})
	defer body.Close()
	return ioutil.ReadAll(body)
}

func NewAzClient() *AzStorage {
	return &AzStorage{}
}
