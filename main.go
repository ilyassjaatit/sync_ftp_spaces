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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT NULL,
		count INTEGER DEFAULT 1,
		ftp_path VARCHAR NOT NULL UNIQUE,
		remote_path VARCHAR,
		type VARCHAR,
		file_hash VARCHAR); 
		UNIQUE KEY index_id (id) USING BTREE,
		UNIQUE KEY index_ftp_path (ftp_path) USING BTREE,
	`

	statement, err := db.Prepare(sqlInit)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlInit)
		os.Remove(dbPath)
		panic("Error create table ftp_synchronizer")
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

		sqlTemp := `INSERT INTO ftp_synchronizer(id,ftp_path, type)
				VALUES(NULL, ?,?) 
		ON CONFLICT(ftp_path) DO UPDATE SET count=count+1, updated_at=CURRENT_TIMESTAMP;
	`
		statement, err := db.Prepare(sqlTemp)
		if err != nil {
			log.Printf("%q: %s\n", err, sqlTemp)
			panic(err)
		}
		_, err = statement.Exec(w.Path(), "NULL")
		if err != nil {
			log.Printf("%q: %s\n", err, sqlTemp)
			panic(err)
		}
	}
}
