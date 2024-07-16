package main

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"log"
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
	log.Printf("Было зачислено средств в размере: %.2f на аккаунт %s. Новый баланс: %.2f\n", amount, a.ID, a.Balance)
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
	log.Printf("Было выведено средств в размере: %.2f с аккаунта %s. Новый баланс: %.2f\n", amount, a.ID, a.Balance)
	return nil
}

// GetBalance implements BankAccount interface
func (a *Account) GetBalance() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	log.Printf("Был проверен баланс для аккаунта %s. Баланс: %.2f\n", a.ID, a.Balance)
	return a.Balance
}

var (
	accounts       = make(map[string]*Account)
	mu             sync.Mutex
	NotEnoughMoney = errors.New("недостаточно средств")
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/accounts", createAccount).Methods("POST")
	r.HandleFunc("/accounts/{id}/deposit", deposit).Methods("POST")
	r.HandleFunc("/accounts/{id}/withdraw", withdraw).Methods("POST")
	r.HandleFunc("/accounts/{id}/balance", getBalance).Methods("GET")

	log.Println("Starting server on :10533")
	log.Fatal(http.ListenAndServe(":10533", r))
}

func createAccount(w http.ResponseWriter, r *http.Request) {
	var acc Account
	if err := json.NewDecoder(r.Body).Decode(&acc); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	mu.Lock()
	accounts[acc.ID] = &acc
	mu.Unlock()
	w.WriteHeader(http.StatusCreated)
	log.Printf("Создан аккаунт %s\n", acc.ID)
}

func deposit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	mu.Lock()
	acc, ok := accounts[id]
	mu.Unlock()
	if !ok {
		http.Error(w, "Аккаунт не найден", http.StatusNotFound)
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
			log.Printf("Ошибка внесения средств %v на аккаунт %s. Ошибка: %s", deposit.Amount, acc.ID, err)
		}
	}()
	w.WriteHeader(http.StatusOK)
}

func withdraw(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	mu.Lock()
	acc, ok := accounts[id]
	mu.Unlock()
	if !ok {
		http.Error(w, "Аккаунт не найден", http.StatusNotFound)
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
		log.Printf("Ошибка вывода средств с аккаунта %s. Ошибка: %s", acc.ID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func getBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	mu.Lock()
	acc, ok := accounts[id]
	mu.Unlock()
	if !ok {
		http.Error(w, "Аккаунт не найден", http.StatusNotFound)
		return
	}
	balance := acc.GetBalance()
	err := json.NewEncoder(w).Encode(map[string]float64{"balance": balance})
	if err != nil {
		log.Printf("Ошибка получения данных баланса для аккаунта %s. Ошибка: %s", id, err)
		return
	}
}
