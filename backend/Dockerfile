# Stage 1: Build the Go app
FROM golang:1.17 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed.
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o main ./cmd

# Stage 2: Run the Go app
FROM debian:buster

# Install dependencies for running migrations
RUN apt-get update && apt-get install -y \
    wget \
    ca-certificates

# Install migrate CLI
RUN wget -O migrate.tar.gz https://github.com/golang-migrate/migrate/releases/download/v4.15.1/migrate.linux-amd64.tar.gz && \
    tar -xvzf migrate.tar.gz && \
    mv migrate.linux-amd64 /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate && \
    rm migrate.tar.gz

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/Makefile ./
COPY --from=builder /app/.env ./

# Run migrations
RUN make up

# Command to run the executable
CMD ["./main"]
