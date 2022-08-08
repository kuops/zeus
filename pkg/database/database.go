package database

import (
	"database/sql"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type database struct {
	db     *gorm.DB
	sqlDB  *sql.DB
	config *Config
}

func NewDatabase(cfg *Config) *database {
	return &database{
		config: cfg,
	}
}

func (db *database) Open() error {
	c := db.config
	rawDB, err := gorm.Open(mysql.Open(c.ConnectionURL()), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed init gorm: %w", err)
	}

	sqlDB, err := rawDB.DB()
	if err != nil {
		return fmt.Errorf("failed get gorm sqlDB: %w", err)
	}

	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(2)
	// 连接重用最大时间
	sqlDB.SetConnMaxLifetime(c.MaxConnectionLifetime)
	// 连接空闲最大时间
	sqlDB.SetConnMaxIdleTime(c.MaxConnectionIdleTime)

	db.db = rawDB
	db.sqlDB = sqlDB
	return nil
}

func (db *database) GetDB() *gorm.DB {
	return db.db
}

func (db *database) Close() error {
	return db.sqlDB.Close()
}
