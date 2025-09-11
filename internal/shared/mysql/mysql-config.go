package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql" // 👈 ESTE IMPORT REGISTRA EL DRIVER
	migrate "github.com/rubenv/sql-migrate"

	"github.com/joho/godotenv"
)

func runMigrations(db *sql.DB) {
	migrations := &migrate.FileMigrationSource{
		Dir: "migrations",
	}

	n, err := migrate.Exec(db, "mysql", migrations, migrate.Up)
	if err != nil {
		log.Fatalf("Could not apply migrations: %v", err)
	}
	log.Printf("Applied %d migrations!\n", n)
}

func NewMySQL() (*sql.DB, error) {
	_ = godotenv.Overload(".env.local", ".env") // prioridad a .env.local; luego .env
	var dsn string
	if os.Getenv("environment") == "local" {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_NAME"),
		)

	} else {
		dsn = os.Getenv("DB_URL")

	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Buenas prácticas de pool
	db.SetMaxOpenConns(20)                  // máximo de conexiones abiertas
	db.SetMaxIdleConns(10)                  // máximo en idle
	db.SetConnMaxLifetime(55 * time.Minute) // reciclar conexiones
	db.SetConnMaxIdleTime(5 * time.Minute)  // tiempo máximo en idle

	// Verificar conexión
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	if os.Getenv("environment") != "local" {
		runMigrations(db)
	}
	return db, nil
}
