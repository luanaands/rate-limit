FROM golang:latest AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

FROM scratch
WORKDIR /app
COPY --from=builder /app/server ./server
COPY cmd/server/.env ./.env
EXPOSE 8080
ENTRYPOINT ["/app/server"]