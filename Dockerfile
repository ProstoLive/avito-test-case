FROM golang:1.25

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o prmanagement main.go

CMD ["./prmanagement"]
