package dbrepo

import (
	"database/sql"
	"github.com/Seician/bookings/internal/config"
	"github.com/Seician/bookings/internal/repository"
)

type mySqlDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}
type testDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

func NewMySqlRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	return &mySqlDBRepo{
		App: a,
		DB:  conn,
	}
}

func NewTestingRepo(a *config.AppConfig) repository.DatabaseRepo {
	return &testDBRepo{
		App: a,
	}
}
