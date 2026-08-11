FROM golang:latest AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

FROM scratch
COPY --from=builder /app/server .
COPY ./cmd/server/.env ./
EXPOSE 8080
ENTRYPOINT ["./server"]