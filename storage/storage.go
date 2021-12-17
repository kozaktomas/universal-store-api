package storage

import (
	"fmt"
	"time"
)

type Entity struct {
	Id      string      `json:"id"`
	Created time.Time   `json:"created"`
	Payload interface{} `json:"payload"`
}

type Storage interface {
	Add(serviceName string, payload interface{}) error
	List(serviceName string) ([]Entity, error)
	Get(serviceName, id string) (Entity, error)
	Delete(serviceName, id string) error
}

func CreateStorageByType(storageType string, serviceNames []string) (Storage, error) {
	switch storageType {
	case "mem":
		return CreateMemStorage(serviceNames), nil
	}

	return nil, fmt.Errorf("unknown storage type %q", storageType)
}
