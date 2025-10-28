package repositories

import (
	"go-api-boilerplate/internal/response"
	"go-api-boilerplate/module/db"
)

type Repository struct {
	DB   *db.DBConnection
	Resp *response.Response
}

func NewRepository(dbc *db.DBConnection, resp *response.Response) *Repository {
	return &Repository{DB: dbc, Resp: resp}
}
