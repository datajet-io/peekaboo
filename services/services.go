package services

import (
	"errors"
)

// Services is a map of services to test
type Services map[string]*Service

// Get returns the service with the given id
func (s *Services) Get(id string) (*Service, error) {

	service, exists := (*s)[id]

	if !exists {
		return nil, errors.New("No service found for the id." + id)
	}

	return service, nil
}
