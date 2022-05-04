package main

import (
	"container/list"
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

func saveFileSQLBase() string {
	sqlTemp := `INSERT INTO ftp_synchronizer(id,ftp_path, type)
				VALUES(NULL, ?,?) 
		ON CONFLICT(ftp_path) DO UPDATE SET count=count+1, updated_at=CURRENT_TIMESTAMP;
	`
	return sqlTemp
}

func saveFilePath(db *sql.DB, w *ftp.Walker) {
	sqlTemp := saveFileSQLBase()
	statement, err := db.Prepare(sqlTemp)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlTemp)
		panic(err)
	}
	_, err = statement.Exec(w.Path(), "dir")
	if err != nil {
		log.Printf("%q: %s\n", err, sqlTemp)
		panic(err)
	}
	time.Sleep(10 * time.Millisecond)
}

func saveDirPath(db *sql.DB, w *ftp.Walker) {
	sqlTemp := saveFileSQLBase()
	statement, err := db.Prepare(sqlTemp)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlTemp)
		panic(err)
	}
	_, err = statement.Exec(w.Path(), "file")
	if err != nil {
		log.Printf("%q: %s\n", err, sqlTemp)
		panic(err)
	}
	time.Sleep(10 * time.Millisecond)
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
	log.Printf("%q: %s\n", err, db)
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
	ftpPassword := os.Getenv("FTP_PASSWORD")
	err = c.Login(ftpUser, ftpPassword)
	if err != nil {
		log.Fatal(err)
	}

	w := c.Walk("/")
	index := 0
	for w.Next() {
		if index == 10 {
			break
		}

		queueDirPach := list.New()
		queueFilesPach := list.New()
		if w.Stat().Type == ftp.EntryTypeFolder {
			queueDirPach.PushBack(w.Path())
			index = index + 1
			continue
		}
		queueFilesPach.PushBack(w.Path())
		index = index + 1
		//saveFilePath(db, w)
		if index == 10 {
			front := queueDirPach.Front()
			fmt.Println(front.Value)
			break
		}

	}
}
