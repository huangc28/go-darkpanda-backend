package image

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"sync"

	"cloud.google.com/go/storage"
	log "github.com/sirupsen/logrus"
)

// TODO: Add timestamp in file name before upload to GCS.

const GCSPublicHost = "storage.googleapis.com"

type GCSEnhancer struct {
	client     *storage.Client
	bucketName string
	linkList   []string
	mutex      *sync.Mutex
}

func NewGCSEnhancer(client *storage.Client, bucketName string) *GCSEnhancer {
	return &GCSEnhancer{
		client:     client,
		bucketName: bucketName,
	}
}

func (e *GCSEnhancer) ObjectLink(attr *storage.ObjectAttrs) string {
	u := url.URL{
		Scheme: "https",
		Host:   GCSPublicHost,
		Path:   fmt.Sprintf("%s/%s", attr.Bucket, attr.Name),
	}

	return u.String()
}

func (e *GCSEnhancer) Upload(ctx context.Context, file io.Reader, uploadFilename string) (string, error) {
	bucket := e.client.Bucket(e.bucketName)
	object := bucket.Object(uploadFilename)
	objwriter := object.NewWriter(ctx)

	if _, err := io.Copy(objwriter, file); err != nil {
		return "", err
	}

	if err := objwriter.Close(); err != nil {
		return "", err
	}

	// ------------------- make the object publicly accessible -------------------
	if err := object.ACL().Set(ctx,
		storage.AllUsers,
		storage.RoleReader); err != nil {

		return "", err
	}

	// ------------------- retrieve object attributes -------------------
	attr, err := object.Attrs(ctx)

	if err != nil {
		return "", err
	}

	// ------------------- combine object link -------------------
	return e.ObjectLink(attr), nil
}

func (e *GCSEnhancer) UploadMultiple(ctx context.Context, headers []*multipart.FileHeader) ([]string, error) {
	quit := make(chan struct{}, 1)
	defer close(quit)

	errChan := make(chan error, 1)
	boolChan := make(chan bool, 1)
	linkChan := make(chan string, 1)

	for _, header := range headers {
		select {
		case <-quit:
			break
		default:
			go func(header *multipart.FileHeader) {
				file, err := header.Open()
				defer file.Close()

				if err != nil {
					boolChan <- false
					errChan <- err
				}

				objectLink, err := e.Upload(ctx, file, filepath.Base(header.Filename))

				if err != nil {
					boolChan <- false
					errChan <- err
				}

				boolChan <- true
				linkChan <- objectLink
			}(header)
		}
	}

	for range headers {
		if <-boolChan == false {
			close(quit)

			return []string{}, <-errChan
		}

		e.linkList = append(e.linkList, <-linkChan)
	}

	log.Infof("All file uploaded success %v", e.linkList)

	return e.linkList, nil
}
