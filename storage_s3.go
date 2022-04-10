package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"os"
	"strings"
	"time"
)

const s3Timeout = 5 * time.Second

type s3Storage struct {
	client     *s3.S3
	memStorage *memStorage
	bucketName string
}

func CreateS3Storage(serviceNames []string) (*s3Storage, error) {
	endpoint := os.Getenv("AWS_S3_ENDPOINT")
	requiredVariables := []string{
		"AWS_ACCESS_KEY",
		"AWS_SECRET_KEY",
		"AWS_BUCKET_NAME",
		"AWS_REGION",
	}

	for _, variable := range requiredVariables {
		if os.Getenv(variable) == "" {
			return nil, fmt.Errorf("could not found environment variable %q for s3 configuration", variable)
		}
	}

	var configs []*aws.Config
	if endpoint != "" {
		configs = append(configs, &aws.Config{Endpoint: &endpoint})
		if !strings.HasPrefix(endpoint, "https") {
			configs = append(configs, &aws.Config{
				DisableSSL: aws.Bool(true),
			})
		}
	}

	configs = append(configs, &aws.Config{
		S3ForcePathStyle: aws.Bool(true),
	})

	sess := session.Must(session.NewSession(configs...))
	s3Client := s3.New(sess)

	storage := &s3Storage{
		client:     s3Client,
		memStorage: CreateMemStorage(serviceNames),
		bucketName: os.Getenv("AWS_BUCKET_NAME"),
	}

	if err := storage.loadServices(serviceNames); err != nil {
		return nil, fmt.Errorf("could not load existing entities from s3 object storage: %w", err)
	}

	return storage, nil
}

// loadServices iterates through all known services and load them
func (storage *s3Storage) loadServices(serviceNames []string) error {
	for _, serviceName := range serviceNames {
		if err := storage.loadService(serviceName); err != nil {
			return err
		}
	}

	return nil
}

// loadService lists service objects on s3 storage and initiate download
func (storage *s3Storage) loadService(serviceName string) error {
	prefix := fmt.Sprintf("%s/", serviceName)
	list, err := storage.client.ListObjects(&s3.ListObjectsInput{
		Bucket: &storage.bucketName,
		Prefix: &prefix,
	})

	if err != nil {
		return fmt.Errorf("could not list s3 objects with prefix %s: %w", prefix, err)
	}

	for _, object := range list.Contents {
		if err = storage.loadEntity(serviceName, *object.Key); err != nil {
			return err
		}
	}

	return nil
}

// loadEntity downloads entity file, decode content and puts entity to memory cache
func (storage *s3Storage) loadEntity(serviceName, objectKey string) error {
	objectOutput, err := storage.client.GetObject(&s3.GetObjectInput{
		Bucket: &storage.bucketName,
		Key:    &objectKey,
	})
	if err != nil {
		return fmt.Errorf("could not download %q object content from s3 object storage, %w", objectKey, err)
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(objectOutput.Body)
	if err != nil {
		return fmt.Errorf("could not read data from %q output object, %w", objectKey, err)
	}

	var e Entity
	err = json.Unmarshal(buf.Bytes(), &e)
	_ = objectOutput.Body.Close()
	if err != nil {
		return fmt.Errorf("could not decode json from %q output object, %w", objectKey, err)
	}

	storage.memStorage.AddEntity(serviceName, e)
	return nil
}

// Add creates new record in memory and uploads entity to s3 object storage
func (storage *s3Storage) Add(serviceName string, payload interface{}) (Entity, error) {
	e, err := createEntity(payload)
	if err != nil {
		return e, err
	}

	data, err := json.Marshal(e)
	if err != nil {
		return Entity{}, fmt.Errorf("could not marshal data for s3 upload: %w", err)
	}

	ctx := context.Background()
	var cancelFn func()
	ctx, cancelFn = context.WithTimeout(ctx, s3Timeout)

	if cancelFn != nil {
		defer cancelFn()
	}

	key := getS3ObjectKey(serviceName, e.Id)
	contentType := "application/json"
	_, err = storage.client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      &storage.bucketName,
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: &contentType,
	})

	if err != nil {
		return Entity{}, fmt.Errorf("could not upload to s3: %w", err)
	}

	storage.memStorage.AddEntity(serviceName, e)

	return e, nil
}

func (storage *s3Storage) List(serviceName string) ([]Entity, error) {
	return storage.memStorage.List(serviceName)
}

func (storage *s3Storage) Get(serviceName, id string) (Entity, error) {
	return storage.memStorage.Get(serviceName, id)
}

func (storage *s3Storage) Delete(serviceName, id string) error {
	ctx := context.Background()
	var cancelFn func()
	ctx, cancelFn = context.WithTimeout(ctx, s3Timeout)

	if cancelFn != nil {
		defer cancelFn()
	}

	key := getS3ObjectKey(serviceName, id)
	_, err := storage.client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: &storage.bucketName,
		Key:    &key,
	})

	if err != nil {
		return fmt.Errorf("could not delete object from s3 bucket: %w", err)
	}

	return storage.memStorage.Delete(serviceName, id)
}

func getS3ObjectKey(serviceName, entityId string) string {
	return fmt.Sprintf("%s/%s.json", serviceName, entityId)
}
