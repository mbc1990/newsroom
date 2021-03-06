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

func (p *PostgresClient) SetScraped(itemId int) {
	sqlStatement := `  
  UPDATE feed_items SET scraped=True WHERE item_id=$1`
	_, err := p.Db.Exec(sqlStatement, itemId)
	if err != nil {
		panic(err)
	}
}

func (p *PostgresClient) GetIdForItem(itemTitle string) int {

	sqlStatement := `
    SELECT item_id FROM feed_items
    WHERE title=$1 
      `
	rows, err := p.Db.Query(sqlStatement, itemTitle)
	defer rows.Close()
	if err != nil {
		panic(err)
	}
	var itemId int
	for rows.Next() {
		if err := rows.Scan(&itemId); err != nil {
			panic(err)
		}
	}
	return itemId
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

func (p *PostgresClient) GetUnscrapedJobs() []ScraperJob {
	sqlStatement := `
    SELECT item_id, link FROM feed_items WHERE scraped=False`
	rows, err := p.Db.Query(sqlStatement)
	defer rows.Close()
	if err != nil {
		panic(err)
	}

	ret := make([]ScraperJob, 0)
	for rows.Next() {
		var id int
		var link string
		if err := rows.Scan(&id, &link); err != nil {
			panic(err)
		}
		job := new(ScraperJob)
		job.ItemId = id
		job.Url = link
		ret = append(ret, *job)
	}
	return ret
}

func (p *PostgresClient) GetScrapedItems() int {
	sqlStatement := `
    SELECT count(*) FROM feed_items WHERE scraped=True`
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

type FeedItem struct {
	Id          int
	Headline    string
	Description string
}

func (p *PostgresClient) GetFeedItems(timespan Timespan) *[]FeedItem {
	sqlStatement := `
    SELECT item_id, title, description FROM feed_items
    WHERE timestamp >= ($1) AND timestamp < ($2)
      `
	rows, err := p.Db.Query(sqlStatement, timespan.Start, timespan.End)
	defer rows.Close()
	if err != nil {
		panic(err)
	}
	var items []FeedItem
	for rows.Next() {
		var itemId int
		var headline string
		var description string
		if err := rows.Scan(&itemId, &headline, &description); err != nil {
			panic(err)
		}
		item := new(FeedItem)
		item.Id = itemId
		item.Headline = headline
		item.Description = description
		items = append(items, *item)
	}
	return &items
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
