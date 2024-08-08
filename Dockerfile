FROM golang:latest

WORKDIR .

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main.go

CMD ["./main.go"]
