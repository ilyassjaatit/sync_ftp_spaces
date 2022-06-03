package main

import (
	"database/sql"
	"fmt"
	"github.com/jlaffaye/ftp"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"io"
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

func savePath(db *sql.DB, path string, file_type ftp.EntryType) {
	sqlTemp := saveFileSQLBase()
	statement, err := db.Prepare(sqlTemp)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlTemp)
		panic(err)
	}
	_, err = statement.Exec(path, file_type)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlTemp)
		panic(err)
	}
}

func firstScan(dbPath string) {
	db, err := sql.Open("sqlite3", dbPath)
	db.SetMaxOpenConns(100)
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
	ch := make(chan ftp.Walker, 10)
	for w.Next() {
		ch <- *w
		if w.Stat().Type == ftp.EntryTypeFolder {
			go saveOrUpdate(db, ch)
		}

		if w.Stat().Type == ftp.EntryTypeFile {
			func() {
				basePath := os.Getenv("SYNC_BASE_PATH")
				res, err := c.Retr(w.Path())
				if err != nil {
					panic(err)
				}
				defer res.Close()
				outFile, err := os.Create(basePath + w.Path())
				if err != nil {
					log.Fatal(err)
				}
				defer outFile.Close()
				_, err = io.Copy(outFile, res)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(basePath + w.Path())
			}()

		}
		if cap(ch) > 9 {
			time.Sleep(10 * time.Millisecond)
		}
	}
	close(ch)
}

func saveOrUpdate(db *sql.DB, c <-chan ftp.Walker) {
	item := <-c
	if item.Stat().Type == ftp.EntryTypeFolder {
		go createDir(item.Path())
	}
	//savePath(db, item.Path(), item.Stat().Type)
}

func createDir(path string) {
	basePath := os.Getenv("SYNC_BASE_PATH")
	_, err := os.Stat(basePath + path)
	if os.IsNotExist(err) {
		err := os.Mkdir(basePath+path, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
	time.Sleep(10 * time.Microsecond)
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
		firstScan(dbPath)
	}

	if err != nil {
		log.Fatal(err)
	}
}
