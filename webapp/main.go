package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	port := os.Getenv("PORT")

	fmt.Printf("server is running on port %s\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux); err != nil {
		fmt.Println(err)

		os.Exit(1)
	}
}
