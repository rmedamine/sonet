package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

var dbPath = "database/" + DATABASE_NAME

func ConnectDatabase() error {
	dbFolder := "./database"
	if _, err := os.Stat(dbFolder); os.IsNotExist(err) {
		// If the directory does not exist, create it
		if err := os.Mkdir(dbFolder, 0o755); err != nil {
			return fmt.Errorf("error creating database folder: %v", err)
		}
		log.Println("Database folder created")
	}
	var err error
	DB, err = sql.Open(DRIVER_NAME, "./database/"+DATABASE_NAME)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	log.Println("Database connected")
	if _, err = DB.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Fatalf("Error setting foreign key support: %v", err)
	}
	return Migrate()
}

func Migrate() error {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting working directory: %v", err)
	}
	// Migrations are in ../migrations/sqlite relative to this file
	migrationsPath := filepath.ToSlash(filepath.Join(wd, "migrations"))
	// Create migrate instance
	fmt.Println(migrationsPath)
	m, err := migrate.New("file://"+migrationsPath, "sqlite3://"+dbPath+"?x-migrations-table=migrations")
	if err != nil {
		err = fmt.Errorf("error creating migration instance: %v", err)
		return err
	}
	// Apply migrations
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		err = fmt.Errorf("error applying migrations: %v", err)
		return err
	}
	fmt.Println("Migrations applied successfully")
	return nil
}
