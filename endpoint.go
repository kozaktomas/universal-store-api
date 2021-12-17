package main

import (
	"fmt"
	"github.com/kozaktomas/universal-store-api/storage"
)

type Service struct {
	Cfg     ServiceConfig
	Storage storage.Storage
}

func (e *Service) Validate(payload map[string]interface{}) error {
	for fieldName, field := range e.Cfg.Fields {
		value, ok := payload[fieldName]
		if err := Validate(field, value, ok); err != nil {
			return fmt.Errorf("field %q - %w", fieldName, err)
		}
	}

	return nil
}

func (e *Service) Put(payload map[string]interface{}) error {
	if err := e.Storage.Add(e.Cfg.Name, payload); err != nil {
		return fmt.Errorf("could not put new entity into storage: %w", err)
	}

	return nil
}

func (e *Service) List() ([]storage.Entity, error) {
	return e.Storage.List(e.Cfg.Name)
}

func (e *Service) Get(id string) (storage.Entity, error) {
	return e.Storage.Get(e.Cfg.Name, id)
}

func (e *Service) Delete(id string) error {
	return e.Storage.Delete(e.Cfg.Name, id)
}
