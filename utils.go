package main

import (
	"github.com/go-pg/pg/v10"
	"github.com/icrowley/fake"
	"net/http"
	"strconv"
)

func populateData(db *pg.DB) error {

	for i := 0; i < 5; i++ {
		w1 := Service{
			Versions:    []int{fake.MonthNum(), fake.Day(), fake.WeekdayNum()},
			Name:        fake.FirstName(),
			Description: fake.SentencesN(3),
		}
		_, err := db.Model(&w1).Insert()
		if err != nil {
			return err
		}
	}
	return nil
}

type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

type QueryOptions struct {
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Name   string `json:"name"`
	Sort   string `json:"sort"`
}

func GetOptions(r *http.Request) QueryOptions {

	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))
	serviceName := query.Get("name")
	sortBy := query.Get("sort")

	if sortBy == "" {
		sortBy = "+created_at"
	}

	if limit <= 0 || limit >= 50 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	return QueryOptions{
		Limit:  limit,
		Offset: offset,
		Name:   serviceName,
		Sort:   sortBy,
	}

}
