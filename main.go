package main

import (
	"fmt"
	"github.com/jlaffaye/ftp"
	"github.com/joho/godotenv"
	"log"
	"os"
	"reflect"
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
	list_files, err := c.List("/")
	if err != nil {
		log.Fatal(err)
	}

	for _, element := range list_files {
		fmt.Println(element)
		fmt.Println(element.Type)
		fmt.Println(element.Name)
		fmt.Println(element.Size)
		fmt.Println(element.Time)
		fmt.Println(element.Target)
		fmt.Println("list_files = ", reflect.TypeOf(element))
	}

}
