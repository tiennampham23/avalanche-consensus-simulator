FROM golang:1.17 as builder
WORKDIR /app
COPY go.mod .
COPY go.sum .

RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server

FROM ubuntu:20.04
RUN mkdir /app
WORKDIR /app
COPY --from=builder /app/server .
CMD ./server server
