package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectionDB() (*gorm.DB, error) {
	cdb := os.Getenv("DATABASE_URL")
	if cdb == "" {
		host := getRequiredEnv("DB_HOST")
		port := getRequiredEnv("DB_PORT")
		user := getRequiredEnv("DB_USER")
		password := getRequiredEnv("DB_PASSWORD")
		dbname := getRequiredEnv("DB_NAME")

		if host == "" || port == "" || user == "" || dbname == "" {
			return nil, fmt.Errorf("missing database configuration: set DATABASE_URL or DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME")
		}

		cdb = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	}

	db, err := gorm.Open(postgres.Open(cdb), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to access database connection: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	return db, nil

}

func getRequiredEnv(key string) string {
	return os.Getenv(key)
}
