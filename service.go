package main

import (
	"fmt"
)

type Service struct {
	Cfg     ServiceConfig
	Storage Storage
}

func (e *Service) Validate(payload map[string]interface{}) error {
	t := true
	return Validate(FieldConfig{
		Name:     "root",
		Type:     "object",
		Required: &t,
		Fields:   &e.Cfg.Fields,
	}, payload, true)
}

func (e *Service) Put(payload map[string]interface{}) error {
	_, err := e.Storage.Add(e.Cfg.Name, payload)
	if err != nil {
		return fmt.Errorf("could not put new entity into storage: %w", err)
	}

	return nil
}

func (e *Service) List() ([]Entity, error) {
	return e.Storage.List(e.Cfg.Name)
}

func (e *Service) Get(id string) (Entity, error) {
	return e.Storage.Get(e.Cfg.Name, id)
}

func (e *Service) Delete(id string) error {
	return e.Storage.Delete(e.Cfg.Name, id)
}
