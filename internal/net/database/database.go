package database

import (
	"fmt"
	"os"
	"sync"
	"time"

	"uptime-go/internal/configuration"
	"uptime-go/internal/net"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Database struct {
	DB    *gorm.DB
	mutex sync.RWMutex
}

func InitializeDatabase() (*Database, error) {
	DBPath := configuration.DBFile

	// Create the directory if it doesn't exist
	if err := os.MkdirAll("/var/uptime-go/db", 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if the database file exists, if not create it
	if _, err := os.Stat(DBPath); os.IsNotExist(err) {
		file, err := os.Create(DBPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create database file: %w", err)
		}
		defer file.Close()
	}

	// Open the database connection using GORM and SQLite with connection pool configuration
	db, err := gorm.Open(sqlite.Open(DBPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Set connection pool settings
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns sets the maximum number of open connections to the database
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused
	sqlDB.SetConnMaxLifetime(time.Hour)

	// SetConnMaxIdleTime sets the maximum amount of time a connection may be idle
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)

	// Migrate the schema
	if err := db.AutoMigrate(
		&net.Monitor{},
		&net.MonitorHistory{},
		&net.Incident{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database schema: %w", err)
	}

	return &Database{DB: db}, nil
}
func (db *Database) UpsertRecord(record any, column string) error {
	// Create record if not exists else update

	db.mutex.Lock()
	defer db.mutex.Unlock()

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: column}},
			UpdateAll: true,
		}).Create(record).Error; err != nil {
			return fmt.Errorf("failed to save record: %w", err)
		}
		return nil
	})
	return err
}

func (db *Database) SaveRecord(record any) error {
	if err := db.DB.Create(record).Error; err != nil {
		return err
	}

	return nil
}

func (db *Database) GetMonitorRecord(url string) *net.Monitor {
	var record net.Monitor

	if err := db.DB.
		Where("url = ?", url).
		Take(&record).Error; err != nil {
		return nil
	}

	return &record
}
