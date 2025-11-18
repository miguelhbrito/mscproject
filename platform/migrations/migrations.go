package migrations

import (
	"database/sql"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"fmt"
)

func InitMigrations(db *sql.DB) {

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		panic(err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	locationMigrations := fmt.Sprintf("file://%s/platform/migrations/", pwd)

	m, err := migrate.NewWithDatabaseInstance(
		locationMigrations,
		"postgres", driver)
	if err != nil {
		panic(err)
	}

	// Only uncomment this for drop and up on everytime you run the application
	//_ = m.Down()
	_ = m.Up()

	fmt.Println("Successfully migrations applied!")
}
