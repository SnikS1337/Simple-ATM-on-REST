package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gorilla/mux"
)

// DBClientInterface defines the methods that a DBClient must implement
type DBClientInterface interface {
	CreateAccount(acc *Account) error
	UpdateBalance(acc *Account) error
	GetBalance(id string) (float64, error)
}

// MockDBClient is a mock implementation of DBClient for testing
type MockDBClient struct{}

func (m *MockDBClient) CreateAccount(acc *Account) error {
	return nil
}

func (m *MockDBClient) UpdateBalance(acc *Account) error {
	return nil
}

func (m *MockDBClient) GetBalance(id string) (float64, error) {
	return accounts[id].Balance, nil
}

func TestAccountOperations(t *testing.T) {
	// Clear global variables before each test
	accounts = make(map[string]*Account)
	mu = sync.Mutex{}

	// Initialize dbClient with a mock implementation
	DBClientInterface := &MockDBClient{}
	if DBClientInterface.CreateAccount(&Account{}) != nil {
		t.Error("Failed to create account")
	}

	// Making router
	r := mux.NewRouter()
	r.HandleFunc("/accounts", createAccount).Methods("POST")
	r.HandleFunc("/accounts/{id}/deposit", deposit).Methods("POST")
	r.HandleFunc("/accounts/{id}/withdraw", withdraw).Methods("POST")
	r.HandleFunc("/accounts/{id}/balance", getBalance).Methods("GET")

	// Create new account
	createAccountReq := Account{ID: "123", Balance: 100.0}
	createAccountReqBody, _ := json.Marshal(createAccountReq)
	req := httptest.NewRequest("POST", "/accounts", bytes.NewBuffer(createAccountReqBody))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Deposit balance
	depositReq := map[string]float64{"amount": 50.0}
	depositReqBody, _ := json.Marshal(depositReq)
	req = httptest.NewRequest("POST", "/accounts/123/deposit", bytes.NewBuffer(depositReqBody))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Check balance
	req = httptest.NewRequest("GET", "/accounts/123/balance", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	var balanceResp map[string]float64
	json.NewDecoder(w.Body).Decode(&balanceResp)
	if balanceResp["balance"] != 150.0 {
		t.Errorf("Expected balance 150.0, got %f", balanceResp["balance"])
	}

	// Withdraw funds
	withdrawReq := map[string]float64{"amount": 75.0}
	withdrawReqBody, _ := json.Marshal(withdrawReq)
	req = httptest.NewRequest("POST", "/accounts/123/withdraw", bytes.NewBuffer(withdrawReqBody))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Check balance after withdraw
	req = httptest.NewRequest("GET", "/accounts/123/balance", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	json.NewDecoder(w.Body).Decode(&balanceResp)
	if balanceResp["balance"] != 75.0 {
		t.Errorf("Expected balance 75.0, got %f", balanceResp["balance"])
	}

	// Attempted withdrawal with insufficient balance
	withdrawReq = map[string]float64{"amount": 100.0}
	withdrawReqBody, _ = json.Marshal(withdrawReq)
	req = httptest.NewRequest("POST", "/accounts/123/withdraw", bytes.NewBuffer(withdrawReqBody))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
