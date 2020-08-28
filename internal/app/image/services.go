package image

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"cloud.google.com/go/storage"
)

const GCSPublicHost = "storage.googleapis.com"

type GCSEnhancer struct {
	client     *storage.Client
	bucketName string
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
