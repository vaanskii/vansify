# Stage 1: Build the Go app
FROM golang:1.23.1 as builder

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

# Install dependencies for running migrations and building the app
RUN apt-get update && apt-get install -y \
    wget \
    ca-certificates \
    tar \
    make  # Install make

# Download and extract migrate CLI
RUN wget -O migrate.tar.gz https://github.com/golang-migrate/migrate/releases/download/v4.15.1/migrate.linux-amd64.tar.gz && \
    tar -xvzf migrate.tar.gz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate && \
    rm migrate.tar.gz

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/Makefile ./

# Define build-time arguments
ARG DB_USER
ARG DB_PASS
ARG DB_HOST
ARG DB_PORT
ARG DB_NAME
ARG JWT_SECRET
ARG SMTP_SERVER
ARG SMTP_PORT
ARG SMTP_USER
ARG SMTP_PASS

# Set environment variables using the ARG values
ENV DB_USER=${DB_USER}
ENV DB_PASS=${DB_PASS}
ENV DB_HOST=${DB_HOST}
ENV DB_PORT=${DB_PORT}
ENV DB_NAME=${DB_NAME}
ENV JWT_SECRET=${JWT_SECRET}
ENV SMTP_SERVER=${SMTP_SERVER}
ENV SMTP_PORT=${SMTP_PORT}
ENV SMTP_USER=${SMTP_USER}
ENV SMTP_PASS=${SMTP_PASS}

# Run migrations
RUN make up

# Command to run the executable
CMD ["./main"]
