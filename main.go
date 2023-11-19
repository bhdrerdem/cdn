package main

import (
	"cdn/api"
	"cdn/storage"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	cloudFrontService := storage.NewCloudFrontService(os.Getenv("DISTRIBUTION_ID"), os.Getenv("DISTRIBUTION_URL"))
	s3Service := storage.NewS3Service(os.Getenv("BUCKET"))

	server := api.NewServer(s3Service, cloudFrontService, os.Getenv("PORT"))

	server.Run()
}
