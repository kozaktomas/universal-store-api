package main

import (
	"fmt"
)

type memStorage struct {
	services map[string][]Entity
}

func CreateMemStorage(serviceNames []string) *memStorage {
	serviceMap := make(map[string][]Entity)

	for _, serviceName := range serviceNames {
		serviceMap[serviceName] = []Entity{}
	}

	return &memStorage{services: serviceMap}
}

func (storage *memStorage) Add(serviceName string, payload interface{}) (Entity, error) {
	e, err := createEntity(payload)
	if err != nil {
		return e, err
	}

	storage.AddEntity(serviceName, e)

	return e, nil
}

func (storage *memStorage) AddEntity(serviceName string, e Entity) {
	storage.services[serviceName] = append(storage.services[serviceName], e)
}

func (storage *memStorage) List(serviceName string) ([]Entity, error) {
	return storage.services[serviceName], nil
}

func (storage *memStorage) Get(serviceName, id string) (Entity, error) {
	for _, entity := range storage.services[serviceName] {
		if entity.Id == id {
			return entity, nil
		}
	}

	return Entity{}, fmt.Errorf("could not find entity with id %q", id)
}

func (storage *memStorage) Delete(serviceName, id string) error {
	// find ID
	idx := 0
	found := false
	for k, entity := range storage.services[serviceName] {
		if entity.Id == id {
			idx = k
			found = true
		}
	}

	if !found {
		return fmt.Errorf("could not find id %q in entity list", id)
	}

	// delete ID from list
	storage.services[serviceName] = append(storage.services[serviceName][:idx], storage.services[serviceName][idx+1:]...)

	return nil
}
