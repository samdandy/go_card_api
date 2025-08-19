package tools

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

type PGDatabase struct {
	db *sql.DB
}

var DB *PGDatabase

func Init() error {
	DB = &PGDatabase{}
	if err := DB.Connect(); err != nil {
		log.Fatalf("Failed to init DB: %v", err)
		return err
	}
	return nil
}

func (p *PGDatabase) Connect() error {
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
	fmt.Printf("Flushing table %s\n", tableName)
	return err
}

func (p *PGDatabase) WriteCardSearchLog(searchCrit string, resultCount int64, wg *sync.WaitGroup) error {
	if wg != nil {
		defer wg.Done()
	}
	query := "INSERT INTO logger.card_search_log (search_term, result_count) VALUES ($1, $2)"
	_, err := p.db.Exec(query, searchCrit, resultCount)
	return err
}

func (p *PGDatabase) CheckUserExists(userName string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM card.user WHERE username=$1)"
	err := p.db.QueryRow(query, userName).Scan(&exists)
	return exists, err
}

func (p *PGDatabase) CreateUser(username string, password string) (int, error) {
	query := "INSERT INTO card.user (username, pw) VALUES ($1, $2) RETURNING id"
	var userID int
	err := p.db.QueryRow(query, username, password).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (p *PGDatabase) GetUserPassword(username string) (string, error) {
	var password string
	query := "SELECT pw FROM card.user WHERE username=$1"
	err := p.db.QueryRow(query, username).Scan(&password)
	return password, err
}

func (p *PGDatabase) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

func (p *PGDatabase) ReadCardSearchLog(wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	query := "SELECT search_term, result_count FROM logger.card_search_log"
	rows, err := p.db.Query(query)
	if err != nil {
		log.Println("Error reading card search log:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var searchTerm string
		var resultCount int64
		if err := rows.Scan(&searchTerm, &resultCount); err != nil {
			log.Println("Error scanning row:", err)
			continue
		}
		fmt.Printf("Search Term: %s, Result Count: %d\n", searchTerm, resultCount)
	}
}
