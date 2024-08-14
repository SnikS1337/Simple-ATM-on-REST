package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"log"
	"log/slog"
	"net/http"
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

	if !useDatabase {
		w.WriteHeader(http.StatusCreated)
		log.Printf("Created account %s\n", acc.ID)
		err := json.NewEncoder(w).Encode(acc)
		if err != nil {
			slog.Info("Error creating account:", err)
		}
		return
	}

	if dbClient == nil {
		http.Error(w, "Database client is not initialized", http.StatusInternalServerError)
		return
	} else {
		err := dbClient.CreateAccount(&acc)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(acc)
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
