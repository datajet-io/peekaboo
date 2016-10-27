package services

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/uber-go/zap"
)

//Services list of services to test
type Services struct {
	Services map[string]*Service
}

//InitServices initalizes a new Services instance
func initServices() Services {

	var srv Services

	srv.Services = make(map[string]*Service, 5)

	return srv

}

func (s *Services) randString() string {

	// Random
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	n := 3 // three letter ids

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// Get returns the service with the given id
func (s *Services) Get(id string) (*Service, error) {

	service, exists := s.Services[id]

	if !exists {
		return nil, errors.New("No service found for the id." + id)
	}

	return service, nil
}

// GetOwnerByCell returns the service with the given id
func (s *Services) GetOwnerByCell(cell string) (*Owner, error) {

	for _, srv := range s.Services {

		for _, owner := range srv.Owners {

			if owner.Cell == cell {
				return &owner, nil
			}

		}
	}

	return nil, errors.New("No owner found with the cell number." + cell)
}

// Add a given service to the list of services to test
func (s *Services) Add(service Service) (*Service, error) {

	// Try 10x times to create new service ID; 10 is a arbritray number of retries
	for i := 0; i < 10; i++ {

		id := s.randString()

		_, err := s.Get(id)

		if err != nil {

			service.ID = id

			s.Services[id] = &service

			return s.Services[id], nil

		}
	}

	return nil, errors.New("Unable add to new service.")
}

//LoadFromFile loads the configuration from the given filepath
func LoadFromFile(filepath string, logger zap.Logger) (*Services, error) {

	data, err := ioutil.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	s := initServices()

	rawServices := []Service{}

	rand.Seed(time.Now().UnixNano())

	if err := json.Unmarshal(data, &rawServices); err != nil {
		return nil, errors.New("Could not read " + filepath + ", unserialization failed.")
	}

	// Put services into map with unique keys which are used later for identifying in messaging

	for _, service := range rawServices {
		service.Logger = logger
		s.Add(service)
	}

	if len(s.Services) == 0 {
		return nil, errors.New("No services found.")
	}

	return &s, nil
}
