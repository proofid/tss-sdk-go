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
		Tenant: os.Getenv("TSS_TENANT"),
		ServerURL: os.Getenv("TSS_SERVER_URL"),
	})

	// If Path is set, use it.
	// Otherwise, if ID is set, use it.
	// Otherwise, use an ID of 1.
	id := 1
	idFromEnv := os.Getenv("TSS_SECRET_ID")
	path := os.Getenv("TSS_SECRET_PATH")
	var err error
	if idFromEnv != "" {
		id, err = strconv.Atoi(idFromEnv)
		if err != nil {
			log.Fatalf("TSS_SECRET_ID must be an integer: %s", err)
			return
		}
	}

	var s *server.Secret
	if path == "" {
		s, err = tss.Secret(id)
	} else {
		s, err = tss.SecretByPath(path)
	}

	if err != nil {
		log.Fatal("Error calling server.Secret", err)
	}

	if pw, ok := s.Field("password"); ok {
		fmt.Print("The password is ", pw)
	}
}
