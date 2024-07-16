# Simple ATM on REST
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![GitHub last commit](https://img.shields.io/github/last-commit/SnikS1337/Simple-ATM-on-REST)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/SnikS1337/Simple-ATM-on-REST)



This project is a simple REST API to simulate the operation of an ATM. It allows you to create accounts, top up your balance, withdraw funds and check your balance.

## Table of Contents

- [Features](#features)
- [Requirements](#requirements)
- [Installation](#installation)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#License)

## Features
- Create a new account
- Deposit money into an account
- Withdraw money from an account
- Check the balance of an account

## Requirements

- Go 1.16 or higher
- Gorilla Mux

## Installation

1. Clone the repository
   ```sh
   git clone https://github.com/SnikS1337/Simple-ATM-on-REST.git
   cd Simple-ATM-on-REST
   
2. Initialize the Go module:
   ```sh
   go mod init github.com/SnikS1337/Simple-ATM-on-REST
   
3. Install dependencies:
   ```sh
   go get -u github.com/gorilla/mux
   
4. Run the server
   ```sh
   go run main.go

## Usage
To use the API, you can use tools like Postman or curl to send HTTP requests to the endpoints.

### **Example requests:**
### Create an Account
`curl -X POST http://localhost:10053/accounts -H "Content-Type: application/json" -d '{"ID": "123", "Balance:" 100.0}'`
### Deposit money
`curl -X POST http://localhost:10533/accounts/123/deposit -H "Content-Type: application/json" -d '{"amount": 50.0}'`
### Withdraw money
`curl -X POST http://localhost:10533/accounts/123/withdraw -H "Content-Type: application/json" -d '{"amount": 30.0}'`
### Check balance
`curl -X GET http://localhost:10533/accounts/123/balance`

## API Endpoints
### Create Account
```
URL: /accounts
Method: POST
Body: {
"ID": "string",
"Balance": "number"
}
Response: 201 Created
```
### Deposit money
```
URL: /accounts/{id}/deposit
Method: POST
Body: {
"amount": "number"
}
Response: 200 OK
```
### Withdraw money
```
URL: /accounts/{id}/withdraw
Method: POST
Body: {
"amount": "number"
}
Response: 200 OK
```
### Checking balance
```
URL: /accounts/{id}/balance
Method: GET
Body: {
"balance": "number"
}
```

## Testing
### To run the tests, use the following command:
```
go test
```

## Contributing
### **Contributions are welcome! Please open an issue or submit a pull request for any changes.**

## License
### This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
