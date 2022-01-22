package main

import (
	"fmt"
	"github.com/kozaktomas/universal-store-api/config"
	"github.com/kozaktomas/universal-store-api/storage"
)

type Service struct {
	Cfg     config.ServiceConfig
	Storage storage.Storage
}

func (e *Service) Validate(payload map[string]interface{}) error {
	t := true
	return Validate(config.FieldConfig{
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

func (e *Service) List() ([]storage.Entity, error) {
	return e.Storage.List(e.Cfg.Name)
}

func (e *Service) Get(id string) (storage.Entity, error) {
	return e.Storage.Get(e.Cfg.Name, id)
}

func (e *Service) Delete(id string) error {
	return e.Storage.Delete(e.Cfg.Name, id)
}
