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

func TestAccountOperations(t *testing.T) {
	// Очищаем глобальные переменные перед каждым тестом
	accounts = make(map[string]*Account)
	mu = sync.Mutex{}

	// Создаем маршрутизатор
	r := mux.NewRouter()
	r.HandleFunc("/accounts", createAccount).Methods("POST")
	r.HandleFunc("/accounts/{id}/deposit", deposit).Methods("POST")
	r.HandleFunc("/accounts/{id}/withdraw", withdraw).Methods("POST")
	r.HandleFunc("/accounts/{id}/balance", getBalance).Methods("GET")

	// Создаем новый аккаунт
	createAccountReq := Account{ID: "123", Balance: 100.0}
	createAccountReqBody, _ := json.Marshal(createAccountReq)
	req := httptest.NewRequest("POST", "/accounts", bytes.NewBuffer(createAccountReqBody))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Пополняем баланс
	depositReq := map[string]float64{"amount": 50.0}
	depositReqBody, _ := json.Marshal(depositReq)
	req = httptest.NewRequest("POST", "/accounts/123/deposit", bytes.NewBuffer(depositReqBody))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Проверяем баланс
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

	// Снимаем средства
	withdrawReq := map[string]float64{"amount": 75.0}
	withdrawReqBody, _ := json.Marshal(withdrawReq)
	req = httptest.NewRequest("POST", "/accounts/123/withdraw", bytes.NewBuffer(withdrawReqBody))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Проверяем баланс после снятия
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

	// Попытка снятия средств с недостаточным балансом
	withdrawReq = map[string]float64{"amount": 100.0}
	withdrawReqBody, _ = json.Marshal(withdrawReq)
	req = httptest.NewRequest("POST", "/accounts/123/withdraw", bytes.NewBuffer(withdrawReqBody))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
