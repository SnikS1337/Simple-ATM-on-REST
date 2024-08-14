package main

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type DBClient struct {
	db *sql.DB
}

func NewDBClient(db *sql.DB) *DBClient {
	return &DBClient{db: db}
}

func (dbc *DBClient) CreateTable() error {
	_, err := dbc.db.Exec("CREATE TABLE IF NOT EXISTS accounts (id serial PRIMARY KEY, balance bigint)")
	if err != nil {
		return err
	}
	return nil
}

func (dbc *DBClient) CreateAccount(acc *Account) error {
	_, err := dbc.db.Exec("INSERT INTO accounts (id, balance) VALUES ($1, $2)", acc.ID, acc.Balance)
	if err != nil {
		return err
	}
	return nil
}

func (dbc *DBClient) UpdateBalance(acc *Account) error {
	_, err := dbc.db.Exec("UPDATE accounts SET balance = $1 WHERE id = $2", acc.Balance, acc.ID)
	if err != nil {
		return err
	}
	return nil
}

func (dbc *DBClient) GetBalance(id string) (float64, error) {
	var balance float64
	err := dbc.db.QueryRow("SELECT balance FROM accounts WHERE id = $1", id).Scan(&balance)
	if err != nil {
		return 0, err
	}
	return balance, nil
}
