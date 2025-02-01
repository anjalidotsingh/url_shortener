package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/anjalidotsingh/url_shortener/route"
	"gopkg.in/yaml.v2"
)

var (
	envVars map[string]string
)

func init() {
	loadEnv()
}

func main() {
	fmt.Println("Server")
	port, exists := envVars["PORT"]
	if !exists {
		log.Fatal("PORT not found in config.yml")
	}
	r := route.SetupRouter()
	http.ListenAndServe(":"+port, r)
}

func loadEnv() {
	data, err := os.ReadFile("config.yml")
	if err != nil {
		log.Fatalf("Failed to read config.yml: %v", err)
	}
	if err := yaml.Unmarshal(data, &envVars); err != nil {
		log.Fatalf("Failed to parse config.yml: %v", err)
	}
}
