FROM golang:latest

WORKDIR /

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./
RUN go build -o main.go

CMD ["./app"]
