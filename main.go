package main

import (
	"database/sql"
	"fmt"
	"github.com/jlaffaye/ftp"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"time"
)

func initDb(dbPath string) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	sqlInit := `CREATE TABLE ftp_synchronizer (
	id INTEGER(11) NOT NULL,
	created_at DATETIME,
	updated_at DATETIME,
	ftp_path VARCHAR NOT NULL,
	remote_path VARCHAR,
	file_hash VARCHAR,
	PRIMARY KEY (id)); `

	statement, err := db.Prepare(sqlInit)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlInit)
		os.Remove(dbPath)
		return
	}
	statement.Exec()
	fmt.Println("OK created Table")
	db.Close()
}

func main() {
	// start sql
	errEnv := godotenv.Load()
	if errEnv != nil {
		log.Fatal("Error loading .env file")
	}

	var dbPath string = "./" + os.Getenv("DB_NAME")
	var dbInit bool = false
	_, err := os.Stat(dbPath)
	if err != nil {
		dbInit = true
		os.Create(dbPath)
	}

	if dbInit == true {
		initDb(dbPath)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(db)
	ftpHost := os.Getenv("FTP_HOST")
	ftpPort := os.Getenv("FTP_PORT")
	c, err := ftp.Dial(ftpHost+":"+ftpPort, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	ftpUser := os.Getenv("FTP_USER")
	ftpPAssword := os.Getenv("FTP_PASSWORD")
	err = c.Login(ftpUser, ftpPAssword)
	if err != nil {
		log.Fatal(err)
	}

	w := c.Walk("/")
	for w.Next() {
		if w.Stat().Type == ftp.EntryTypeFolder {
			continue
		}
		fmt.Println(w.Path())
	}
}
