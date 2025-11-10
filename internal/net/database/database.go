package database

import (
	"fmt"
	"os"
	"sync"
	"time"
	"uptime-go/internal/incident"
	"uptime-go/internal/models"

	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	instance *Database
	once     sync.Once
)

type Database struct {
	DB    *gorm.DB
	mutex sync.RWMutex
}

func Get() *Database {
	return instance
}

func Init(dbPath string) (err error) {
	once.Do(func() {
		// Check if the database file exists, if not create it
		if _, errStat := os.Stat(dbPath); dbPath != ":memory:" && err != nil {
			if !os.IsNotExist(errStat) {
				err = errStat
				return
			}

			file, errCreate := os.Create(dbPath)
			if errCreate != nil {
				err = fmt.Errorf("failed to create database file: %w", errCreate)
			}
			file.Close()
		}

		log.Debug().Str("database", dbPath).Msg("connectiong to database...")

		// Open the database connection using GORM and SQLite with connection pool configuration
		gormDB, errOpen := gorm.Open(sqlite.Open(dbPath+"?_journal_mode=WAL&_pragma=foreign_keys"), &gorm.Config{})
		if errOpen != nil {
			err = fmt.Errorf("failed to connect to database: %w", errOpen)
			return
		}

		// Configure connection pooling
		sqlDB, errSQL := gormDB.DB()
		if errSQL != nil {
			err = fmt.Errorf("failed to get database connection: %w", errSQL)
			return
		}

		// Set connection pool settings
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
		sqlDB.SetConnMaxIdleTime(30 * time.Minute)

		// Migrate the schema
		if errMigrate := gormDB.AutoMigrate(
			&models.Monitor{},
			&models.MonitorHistory{},
			&models.Incident{},
		); errMigrate != nil {
			err = fmt.Errorf("failed to migrate database schema: %w", errMigrate)
			return
		}

		instance = &Database{DB: gormDB}
	})

	return err
}

func InitializeTestDatabase() (*Database, error) {
	db, err := gorm.Open(sqlite.Open(":memory:?_journal_mode=WAL&_pragma=foreign_keys"), &gorm.Config{
		// NamingStrategy: schema.NamingStrategy{
		// 	TablePrefix: "test_" + fmt.Sprint(time.Now().Unix()) + "_",
		// },
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.AutoMigrate(
		&models.Monitor{},
		&models.MonitorHistory{},
		&models.Incident{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database schema: %w", err)
	}

	return &Database{DB: db}, nil
}

func (db *Database) UpsertRecord(record any, column string, updateColumn *[]string) error {
	// Create record if not exists else update

	db.mutex.Lock()
	defer db.mutex.Unlock()

	stmt := clause.OnConflict{
		Columns:   []clause.Column{{Name: column}},
		UpdateAll: true,
	}

	if updateColumn != nil {
		stmt.UpdateAll = false
		stmt.DoUpdates = clause.AssignmentColumns(*updateColumn)
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(stmt).Create(record).Error; err != nil {
			return fmt.Errorf("failed to save record: %w", err)
		}
		return nil
	})
}

func (db *Database) Upsert(record any) error {
	return db.UpsertRecord(record, "id", nil)
}

func (db *Database) GetLastIncident(url string, incidentType incident.Type) *models.Incident {
	var incident models.Incident

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	db.DB.Joins("Monitor").
		Select("incidents.*").
		Where("Monitor.url = ? AND incidents.type = ? AND incidents.solved_at IS NULL", url, incidentType).
		Order("incidents.created_at DESC").
		Limit(1).
		Find(&incident)

	return &incident
}
