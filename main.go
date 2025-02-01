package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/anjalidotsingh/url_shortener/route"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
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
	dbconn, err := getDBConnection()
	if err != nil {
		log.Fatalf("Database initialization error: %v", err)
	}
	if dbconn != nil {
		defer dbconn.Close()
	}
	// tx, err := dbconn.Beginx()
	// if err != nil {
	// 	log.Fatalf("Database initialization error: %v", err)
	// }
	handler := route.Handler{
		DB: dbconn,
	}
	r := handler.SetupRouter()
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

// GetDBConnection initializes and returns a database connection
func getDBConnection() (*sqlx.DB, error) {
	username := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	if username == "" || password == "" || host == "" || port == "" || dbName == "" {
		return nil, fmt.Errorf("database environment variables are not set")
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", username, password, host, port, dbName)
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}
