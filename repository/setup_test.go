package repository

import (
	"chat/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates an isolated in-memory SQLite database using a temporary
// file and runs auto-migrations for all models. Each call returns a completely
// fresh, isolated database so tests never share state.
func setupTestDB() *gorm.DB {
	// Use plain ":memory:" (no cache=shared) so each GORM instance gets its own
	// private in-memory database, fully isolated from other calls to setupTestDB.
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic("failed to connect to in-memory database: " + err.Error())
	}

	// Limit to a single connection so the in-memory database is consistent.
	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to get underlying sql.DB: " + err.Error())
	}
	sqlDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(
		&models.User{},
		&models.Group{},
		&models.GroupMember{},
		&models.Message{},
		&models.Seen{},
		&models.File{},
	); err != nil {
		panic("failed to run migrations: " + err.Error())
	}

	return db
}
