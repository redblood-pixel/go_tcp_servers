FROM golang:alpine

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -tags=docker -o /main

CMD ["/main"]