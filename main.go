package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/thycotic/tss-sdk-go/server"
)

func main() {
	tss, _ := server.New(server.Configuration{
		Credentials: server.UserCredential{
			Username: os.Getenv("TSS_USERNAME"),
			Password: os.Getenv("TSS_PASSWORD"),
		},
		// Expecting either the tenant or URL to be set
		Tenant:    os.Getenv("TSS_TENANT"),
		ServerURL: os.Getenv("TSS_SERVER_URL"),
	})

	id, err := strconv.Atoi(os.Getenv("TSS_SECRET_ID"))
	if err == nil {
		s, err := tss.Secret(id)
		if err != nil {
			log.Fatal("Error calling server.Secret by id: ", err)
		}
		if pw, ok := s.Field("password"); ok {
			fmt.Printf("The password for id '%d' is %s\n", id, pw)
		}
	}

	pathAndName := os.Getenv("TSS_SECRET_PATH")
	if pathAndName != "" {
		s, err := tss.Secret(pathAndName)
		if err != nil {
			log.Fatal("Error calling server.Secret by folder path and name: ", err)
		}
		if pw, ok := s.Field("password"); ok {
			fmt.Printf("The password at '%s' is %s\n", pathAndName, pw)
		}
	}
}
