package main

import (
	"banking/router"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	password := os.Getenv("PASSWORD")
	if password == "" {
		log.Fatal("PASSWORD environment variable is not set")
	}

	fmt.Println("MongoDB API")
	r := routes.Routers(password)
	fmt.Println("Server is getting started...")
	log.Fatal(http.ListenAndServe(":"+port, r))
	fmt.Println("Litsening on Port", port)
} 
