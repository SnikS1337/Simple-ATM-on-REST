package main

import (
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	_ "github.com/swaggo/http-swagger/example/go-chi/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// main make entry point, start the server and initialize methods
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	scanner()

	r := mux.NewRouter()

	r.HandleFunc("/accounts", createAccount).Methods("POST")
	r.HandleFunc("/accounts/{id}/deposit", deposit).Methods("POST")
	r.HandleFunc("/accounts/{id}/withdraw", withdraw).Methods("POST")
	r.HandleFunc("/accounts/{id}/balance", getBalance).Methods("GET")

	r.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	r.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Could not get working directory: %s", err)
		}
		yamlFilePath, err := filepath.Rel(wd, "swagger.yaml")
		if err != nil {
			http.Error(w, "Could not find Swagger YAML file", http.StatusInternalServerError)
			return
		}
		yamlFile, err := os.Open(yamlFilePath)
		if err != nil {
			http.Error(w, "Could not open Swagger YAML file", http.StatusInternalServerError)
			return
		}
		defer yamlFile.Close()

		w.Header().Set("Content-Type", "application/json")
		fileInfo, err := os.Stat(yamlFilePath)
		if err != nil {
			http.Error(w, "Could not get file info", http.StatusInternalServerError)
			return
		}
		http.ServeContent(w, r, "doc.json", fileInfo.ModTime(), yamlFile)
	})

	handler := cors.Default().Handler(r)
	log.Println("Starting server on port: 8080")
	log.Fatal(http.ListenAndServe(":8080", handler))

}
