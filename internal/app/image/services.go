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
)

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
	for _, header := range headers {
		//log.Printf("DEBUG 8 %s", header.Filename)
		file, err := header.Open()
		defer file.Close()

		if err != nil {
			return []string{}, err
		}

		objectLink, err := e.Upload(ctx, file, filepath.Base(header.Filename))
		if err != nil {
			return []string{}, err
		}

		e.linkList = append(e.linkList, objectLink)
	}

	return e.linkList, nil
}
