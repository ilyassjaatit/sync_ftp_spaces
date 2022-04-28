package main

import (
	"fmt"
	"github.com/jlaffaye/ftp"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
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
		fmt.Println(w.Path())
	}
}
