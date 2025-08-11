package tools

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type Card struct {
	Name  string
	Price int64
}

type PGDatabase struct {
	db *sql.DB
}

type Database interface {
	InitDB() error
	FlushTable(tableName string) error
	WriteCardSearchLog(searchCrit string, resultCount int64) error
	Close() error
}

var DB *PGDatabase

func Init() {
	DB = &PGDatabase{}
	if err := DB.InitDB(); err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}
}

func (p *PGDatabase) InitDB() error {
	host := os.Getenv("PG_HOST")
	port := os.Getenv("PG_PORT")
	user := os.Getenv("PG_USER")
	password := os.Getenv("PG_PASSWORD")
	dbname := os.Getenv("PG_DBNAME")
	sslmode := os.Getenv("PG_SSLMODE")
	if port == "" {
		port = "5432"
	}
	if sslmode == "" {
		sslmode = "disable"
	}
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)
	var err error
	p.db, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	return p.db.Ping()
}

func (p *PGDatabase) FlushTable(tableName string) error {
	query := fmt.Sprintf("DELETE FROM logger.%s;", tableName)
	_, err := p.db.Exec(query)
	return err
}

func (p *PGDatabase) WriteCardSearchLog(searchCrit string, resultCount int64) error {
	query := "INSERT INTO logger.card_search_log (search_term, result_count) VALUES ($1, $2)"
	_, err := p.db.Exec(query, searchCrit, resultCount)
	return err
}

func (p *PGDatabase) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}
