package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

const createdKey = "usa-internal-created"

type firestoreStorage struct {
	client           *firestore.Client
	collectionPrefix string
}

func CreateFirestoreStorage() (*firestoreStorage, error) {
	googleProjectIdEnv := "GOOGLE_PROJECT_ID"
	googleProjectId := os.Getenv(googleProjectIdEnv)
	if googleProjectId == "" {
		return nil, fmt.Errorf("could not find environment variable %q for firestore configuration", googleProjectIdEnv)
	}

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, googleProjectId)

	if value := os.Getenv("FIRESTORE_EMULATOR_HOST"); value != "" {
		log.Printf("Using Firestore Emulator: %s", value)
	}

	if err != nil {
		return nil, fmt.Errorf("could not connect to firestore: %w", err)
	}

	return &firestoreStorage{
		client:           client,
		collectionPrefix: os.Getenv("GOOGLE_FIRESTORE_COLLECTION_PREFIX"),
	}, nil
}

func (fs *firestoreStorage) Add(serviceName string, payload interface{}) (Entity, error) {
	e, err := createEntity(payload)
	if err != nil {
		return e, err
	}

	// store created value as part of payload
	data, _ := payload.(map[string]interface{})
	data[createdKey] = e.Created

	cn := fs.getCollectionName(serviceName)
	ctx := context.Background()
	_, err = fs.client.Collection(cn).Doc(e.Id).Set(ctx, payload)
	if err != nil {
		return Entity{}, fmt.Errorf("could not write data to firestore: %w", err)
	}

	return e, nil
}

func (fs *firestoreStorage) List(serviceName string) ([]Entity, error) {
	ctx := context.Background()
	cn := fs.getCollectionName(serviceName)
	docs, err := fs.client.Collection(cn).
		Documents(ctx).
		GetAll()
	if err != nil {
		return []Entity{}, fmt.Errorf("could not list items from firestore: %w", err)
	}

	entities := []Entity{}
	for _, doc := range docs {
		data := doc.Data()
		created, ok := data[createdKey].(time.Time)
		if !ok {
			return []Entity{}, fmt.Errorf("could not decode created value")
		}
		delete(data, createdKey)

		entities = append(entities, Entity{
			Id:      doc.Ref.ID,
			Created: created,
			Payload: doc.Data(),
		})
	}

	return entities, nil
}

func (fs *firestoreStorage) Get(serviceName, id string) (Entity, error) {
	ctx := context.Background()
	cn := fs.getCollectionName(serviceName)
	doc, err := fs.client.Collection(cn).Doc(id).Get(ctx)
	if err != nil {
		return Entity{}, fmt.Errorf("could not find entity %q: %w", id, err)
	}

	data := doc.Data()
	created, ok := data[createdKey].(time.Time)
	if !ok {
		return Entity{}, fmt.Errorf("could not decode created value")
	}
	delete(data, createdKey)

	return Entity{
		Id:      doc.Ref.ID,
		Created: created,
		Payload: data,
	}, nil
}

func (fs *firestoreStorage) Delete(serviceName, id string) error {
	ctx := context.Background()
	cn := fs.getCollectionName(serviceName)
	_, err := fs.client.Collection(cn).Doc(id).Delete(ctx)

	if err != nil {
		return fmt.Errorf("could not delele entity %s: %w", id, err)
	}

	return nil
}

func (fs *firestoreStorage) getCollectionName(serviceName string) string {
	return fmt.Sprintf("%s%s", fs.collectionPrefix, serviceName)
}
