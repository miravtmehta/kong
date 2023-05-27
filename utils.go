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
			Description: fake.SentencesN(2),
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
		sortBy = "created_at"
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

func optionsHelper(opts *QueryOptions) (string, string, error) {
	order := "ASC"
	orderBy := opts.Sort[:]
	if opts.Sort[0] == '-' {
		order = "DESC"
		orderBy = opts.Sort[1:]
	} else {
		order = "ASC"
	}
	return orderBy, order, nil
}
