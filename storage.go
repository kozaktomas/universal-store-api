package main

import (
	"fmt"
	uuid "github.com/nu7hatch/gouuid"
	"time"
)

type Entity struct {
	Id      string      `json:"id"`
	Created time.Time   `json:"created"`
	Payload interface{} `json:"payload"`
}

func createEntity(payload interface{}) (Entity, error) {
	uuidV4, err := uuid.NewV4()
	if err != nil {
		return Entity{}, fmt.Errorf("could not create new ID for entity: %w", err)
	}

	return Entity{
		Id:      uuidV4.String(),
		Created: time.Now(),
		Payload: payload,
	}, nil
}

type Storage interface {
	Add(serviceName string, payload interface{}) (Entity, error)
	List(serviceName string) ([]Entity, error)
	Get(serviceName, id string) (Entity, error)
	Delete(serviceName, id string) error
}

func CreateStorageByType(storageType string, serviceNames []string) (Storage, error) {
	switch storageType {
	case "mem":
		return CreateMemStorage(serviceNames), nil
	case "s3":
		return CreateS3Storage(serviceNames)
	}

	return nil, fmt.Errorf("unknown storage type %q", storageType)
}
