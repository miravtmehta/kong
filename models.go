package main

import (
	"fmt"
	"sort"
)

func (s *serviceClient) GetAllService(opts *QueryOptions) (*ServiceList, error) {
	var dbModel []Service
	orderBy, order, err := optionsHelper(opts)
	if err != nil {
		return nil, err
	}
	count, err := s.client.
		Model(&dbModel).
		Where("name ILIKE ?", "%"+opts.Name+"%").
		Limit(opts.Limit).
		Offset(opts.Offset).
		Order(fmt.Sprintf("%s %s", orderBy, order)).
		SelectAndCount()

	if err != nil {
		return nil, err
	}
	for i := 0; i < len(dbModel); i++ {
		sort.Ints(dbModel[i].Versions)
	}
	return &ServiceList{
		Result: dbModel,
		Metadata: Pagination{
			Limit:  opts.Limit,
			Offset: opts.Offset,
			Total:  count,
		},
	}, nil
}

func (s *serviceClient) GetService(name string) (*Service, error) {
	var dbModel Service
	err := s.client.
		Model(&dbModel).
		Where("name = ?", name).Select()
	if err != nil {
		return nil, err
	}
	sort.Ints(dbModel.Versions)
	return &dbModel, nil
}

func (s *serviceClient) CreateService(service Service) error {

	serviceNameExist, err := s.GetService(service.Name)
	if err != nil {
		if _, errIns := s.client.Model(&service).Insert(); errIns != nil {
			return errIns
		}
	}
	if serviceNameExist.Name == service.Name && serviceNameExist.Description == service.Description {
		return fmt.Errorf("service name with %s already exists", service.Name)
	}
	return nil
}

func (s *serviceClient) DeleteService(name string) error {
	var dbModel Service

	// clean all rows
	if name == "*" {
		_, err := s.client.Model(&dbModel).Where("TRUE").Delete()
		if err != nil {
			return err
		}
		return nil
	}
	_, err := s.GetService(name)
	if err != nil {
		return err
	}
	if _, errDel := s.client.Model(&dbModel).Where("name = ?", name).Delete(); errDel != nil {
		return errDel
	}
	return nil
}

func (s *serviceClient) GenerateRandomPgData() error {
	if err := populateData(s.client); err != nil {
		return err
	}
	return nil
}
