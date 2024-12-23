# Use an official Golang image as the base
FROM golang:1.23

# Set environment variables for Go
ENV GO111MODULE=on
ENV CGO_ENABLED=1

# Install required OS dependencies and librdkafka
RUN apt-get update && apt-get install -y \
    software-properties-common \
    build-essential \
    wget \
    && wget -qO - https://packages.confluent.io/deb/7.0/archive.key | apt-key add - \
    && add-apt-repository "deb [arch=amd64] https://packages.confluent.io/deb/7.0 stable main" \
    && apt-get update && apt-get install -y librdkafka-dev

# Set the working directory
WORKDIR /app

# Copy the Go modules manifest and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy application code
COPY . .

# Build the application binary
RUN go build -o main .

# Expose ports if needed (e.g., for HTTP services)
EXPOSE 8080

# Run the application
CMD ["./main"]
