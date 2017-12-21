package main

import "database/sql"
import "fmt"
import _ "github.com/lib/pq"

// Wrapper around postgres interactions
type PostgresClient struct {
	Host     string
	Port     int
	User     string
	Password string
	Dbname   string
	Db       *sql.DB
}

// Add feed item to the feed items table
func (p *PostgresClient) InsertFeedItem(feedTitle string, title string, content string,
	description string, link string) {
	sqlStatement := `  
  INSERT INTO feed_items (feed_title, title, content, description, link, scraped)
  VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING`
	_, err := p.Db.Exec(sqlStatement, feedTitle, title, content, description, link, false)
	if err != nil {
		panic(err)
	}
}

func (p *PostgresClient) GetNumFeedItems() int {
	sqlStatement := `
    SELECT count(*) FROM feed_items`
	rows, err := p.Db.Query(sqlStatement)
	defer rows.Close()
	if err != nil {
		panic(err)
	}

	var count int
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			panic(err)
		}
	}
	return count
}

func (p *PostgresClient) GetDB() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		p.Host, p.Port, p.User, p.Password, p.Dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	return db
}

func NewPostgresClient(pgHost string, pgPort int, pgUser string,
	pgPassword string, pgDbname string) *PostgresClient {
	p := new(PostgresClient)
	p.Host = pgHost
	p.Port = pgPort
	p.User = pgUser
	p.Password = pgPassword
	p.Dbname = pgDbname
	p.Db = p.GetDB()
	p.Db.SetMaxOpenConns(50)
	return p
}
