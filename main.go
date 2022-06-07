package main

import (
	"fmt"
	"github.com/jlaffaye/ftp"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"os"
	"time"
)

func firstScan() {
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
	for w.Next() {
		fmt.Println(w.Path())

		if w.Stat().Type == ftp.EntryTypeFile {
			func() {
				basePath := os.Getenv("SYNC_BASE_PATH")
				res, err := c.Retr(w.Path())
				if err != nil {
					panic(err)
				}
				outFile, err := os.Create(basePath + w.Path())
				if err != nil {
					log.Fatal(err)
				}
				_, err = io.Copy(outFile, res)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(basePath + w.Path())
				res.Close()
				outFile.Close()
			}()
		}
	}
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
	firstScan()
}
