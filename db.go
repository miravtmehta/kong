package main

import (
	"context"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"time"
)

type Service struct {
	Id          int       `pg:",pk" pg:"auto_increment" json:"id"`
	Versions    []int     `json:"versions"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `pg:"default:now()" json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

type ServiceList struct {
	Result   []Service  `json:"result,omitempty"`
	Metadata Pagination `json:"metadata,omitempty"`
}

type serviceClient struct {
	client *pg.DB
}

var _ ServiceProvider = (*serviceClient)(nil)

func NewServiceClient(db *pg.DB) ServiceProvider {
	return &serviceClient{client: db}

}
func NewPostgresConnection(ctx context.Context) (*pg.DB, error) {
	db := pg.Connect(
		&pg.Options{
			Addr:     "localhost:5432",
			User:     "postgres",
			Password: "passo",
		})

	err := db.WithContext(ctx).Ping(ctx)
	if err != nil {
		return nil, err
	}

	err = createSchema(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func createSchema(db *pg.DB) error {
	models := []interface{}{
		(*Service)(nil),
	}
	for _, model := range models {
		err := db.Model(model).CreateTable(&orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
