package main

type ServiceProvider interface {
	GetAllService(pagination *QueryOptions) (*ServiceList, error)
	GetService(name string) (*Service, error)
	CreateService(service Service) error
	DeleteService(name string) error
	GenerateRandomPgData() error
}
