package database

import (
	"context"
	"database/sql"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Open(ctx context.Context, dsn string) (*gorm.DB, *sql.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		return nil, nil, err
	}

	sqldb, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	if err := sqldb.PingContext(ctx); err != nil {
		_ = sqldb.Close()
		return nil, nil, err
	}

	return db, sqldb, nil
}
