package main

import (
	"fmt"
	"sort"
)

func (s *serviceClient) GetAllService(opts *QueryOptions) (*ServiceList, error) {
	var dbModel []Service
	var order string

	if opts.Sort[0] == '+' {
		order = "ASC"
	} else if opts.Sort[0] == '-' {
		order = "DESC"
	} else {
		return nil, fmt.Errorf("invalid sort option %s", opts.Sort)
	}

	orderBy := opts.Sort[1:]
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
	_, err := s.client.Model(&service).Insert()
	if err != nil {
		return err
	}
	return nil
}

func (s *serviceClient) DeleteService(name string) error {
	var dbModel Service
	_, err := s.client.
		Model(&dbModel).
		Where("name = ?", name).Delete()
	if err != nil {
		return err
	}
	return nil
}

func (s *serviceClient) GenerateRandomPgData() error {
	err := populateData(s.client)
	if err != nil {
		return err
	}
	return nil
}
