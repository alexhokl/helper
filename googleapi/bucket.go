package googleapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"cloud.google.com/go/storage"
	google "google.golang.org/api/googleapi"
)

func IsBucketAccessible(ctx context.Context, storageClient *storage.Client, bucketName string) (bool, error) {
	_, err := storageClient.Bucket(bucketName).Attrs(ctx)
	if err != nil {
		if err == storage.ErrBucketNotExist {
			return false, nil
		}
		var e *google.Error
		if ok := errors.As(err, &e); ok && e.Code == http.StatusForbidden {
			return false, nil
		}
		return false, fmt.Errorf("unable to retrieve attributes of bucket [gs://%s]: %w", bucketName, err)
	}
	return true, nil
}

func IsBucketObjectExist(ctx context.Context, storageClient *storage.Client, bucketName string, objectPath string) (bool, error) {
	_, err := getBucketObjectAttributes(ctx, storageClient, bucketName, objectPath)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return false, nil
		}
		return false, fmt.Errorf("unable to retrieve attributes of object [%s] in bucket [gs://%s]: %w", objectPath, bucketName, err)
	}
	return true, nil
}

func GetBucketObjectChecksum(ctx context.Context, storageClient *storage.Client, bucketName string, objectPath string) (uint32, error) {
	attrs, err := getBucketObjectAttributes(ctx, storageClient, bucketName, objectPath)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return 0, fmt.Errorf("object [%s] does not exist in bucket [gs://%s]", objectPath, bucketName)
		}
		return 0, fmt.Errorf("unable to retrieve attributes of object [%s] in bucket [gs://%s]: %w", objectPath, bucketName, err)
	}
	return attrs.CRC32C, nil
}

func getBucketObjectAttributes(ctx context.Context, storageClient *storage.Client, bucketName string, objectPath string) (*storage.ObjectAttrs, error) {
	return storageClient.Bucket(bucketName).Object(objectPath).Attrs(ctx)
}

