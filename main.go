package main

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/jlaffaye/ftp"
)

func main() {
	c, err := ftp.Dial("test_server_ftp", ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Fatal(err)
	}

	err = c.Login("test_user", "test_password")
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
