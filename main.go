package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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
	"sync"
)

// BankAccount interface
type BankAccount interface {
	Deposit(amount float64) error
	Withdraw(amount float64) error
	GetBalance() float64
}

// Account struct
type Account struct {
	ID      string
	Balance float64
	mu      sync.Mutex
}

// Deposit implements BankAccount interface
func (a *Account) Deposit(amount float64) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Balance += amount
	log.Printf("Deposited: %.2f to account %s. New balance: %.2f\n", amount, a.ID, a.Balance)
	return nil
}

// Withdraw implements BankAccount interface
func (a *Account) Withdraw(amount float64) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.Balance < amount {
		return NotEnoughMoney
	}
	a.Balance -= amount
	log.Printf("Withdrawed: %.2f from account %s. New balance: %.2f\n", amount, a.ID, a.Balance)
	return nil
}

// GetBalance implements BankAccount interface
func (a *Account) GetBalance() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	log.Printf("Checked balance for account %s. Balance: %.2f\n", a.ID, a.Balance)
	return a.Balance
}

var (
	db             *sql.DB
	dbClient       *DBClient
	accounts       = make(map[string]*Account)
	mu             sync.Mutex
	NotEnoughMoney = errors.New("not enough money")
)

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

//TO:DO 136 STR AUTH

//type (
//	AuthHandler struct {
//		storage *AuthStorage
//	}
//
//	AuthStorage struct {
//		users map[string]User
//	}
//
//	User struct {
//		ID      string
//		Balance float64
//	}
//)
//
//type RegisterRequest struct {
//	ID      string  `json:"id"`
//	Balance float64 `json:"balance"`
//}
//
//func (h *AuthHandler) Register(c *fiber.Ctx) error {
//	regReq := RegisterRequest{}
//	if err := c.BodyParser(&regReq); err != nil {
//		return fmt.Errorf("could not parse body: %w", err)
//	}
//
//	if _, ok := h.storage.users[regReq.ID]; ok {
//		return errors.New("user already exists")
//	}
//
//	h.storage.users[regReq.ID] = User{
//		ID:      regReq.ID,
//		Balance: regReq.Balance,
//	}
//
//	return c.SendStatus(fiber.StatusCreated)
//
//}

// main make entry point, start the server and initialize methods
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err = sql.Open("postgres", "postgres://postgres:admin@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		log.Fatal("Error while connecting to database", err)
	}
	defer db.Close()

	dbClient = NewDBClient(db)

	r := mux.NewRouter()

	//app := fiber.New()
	//authHandler := &AuthHandler{&AuthStorage{map[string]User{}}} 			//TO:DO add auth
	//app.Post("/register", authHandler.Register)
	//logrus.Fatal(app.Listen(":80"))

	r.HandleFunc("/accounts", createAccount).Methods("POST")
	r.HandleFunc("/accounts/{id}/deposit", deposit).Methods("POST")
	r.HandleFunc("/accounts/{id}/withdraw", withdraw).Methods("POST")
	r.HandleFunc("/accounts/{id}/balance", getBalance).Methods("GET")

	r.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	r.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		yamlFilePath, err := filepath.Abs("D:\\GOLANG\\BANKOMAT\\Simple-ATM-on-REST\\swagger.yaml")
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
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// createAccount creates new account
func createAccount(w http.ResponseWriter, r *http.Request) {
	var acc Account
	if err := json.NewDecoder(r.Body).Decode(&acc); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	mu.Lock()
	accounts[acc.ID] = &acc
	mu.Unlock()
	if dbClient == nil {
		http.Error(w, "Database client is not initialized", http.StatusInternalServerError)
		return
	}
	err := dbClient.CreateAccount(&acc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	log.Printf("Created account %s\n", acc.ID)
}

// deposit funds to account
func deposit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	mu.Lock()
	acc, ok := accounts[id]
	mu.Unlock()
	if !ok {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}
	var deposit struct {
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&deposit); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	go func() {
		err := acc.Deposit(deposit.Amount)
		if err != nil {
			log.Printf("Error deposit funds %v to account %s. Error: %s", deposit.Amount, acc.ID, err)
		}
		err = dbClient.UpdateBalance(acc)
		if err != nil {
			log.Printf("Error updating balance for account %s. Error: %s", acc.ID, err)
		}
	}()
	w.WriteHeader(http.StatusOK)
}

// withdraw funds from account
func withdraw(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	mu.Lock()
	acc, ok := accounts[id]
	mu.Unlock()
	if !ok {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}
	var withdrawal struct {
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&withdrawal); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := acc.Withdraw(withdrawal.Amount); err != nil {
		log.Printf("Error withdraw funds from account %s. Error: %s", acc.ID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err := dbClient.UpdateBalance(acc)
	if err != nil {
		log.Printf("Error updating balance in DB for account %s. Error: %s", acc.ID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// getBalance returns the account balance
func getBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	mu.Lock()
	acc, ok := accounts[id]
	mu.Unlock()
	if !ok {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}
	balance := acc.GetBalance()
	err := json.NewEncoder(w).Encode(map[string]float64{"balance": balance})
	if err != nil {
		log.Printf("Error getting balance for account %s. Error: %s", id, err)
		return
	}
}

func getDBConnection() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)
}
