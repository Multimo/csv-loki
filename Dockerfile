# Use the official Go image as a base
FROM golang:1.21 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go application and CSV file into the container
COPY main.go .
COPY Prod1.csv .

# Build the Go binary
RUN go mod init csvlogger && go mod tidy && go build -o process_csv

# Final image for running the application
FROM debian:bullseye-slim

# Create directories for application and logs
WORKDIR /app
RUN mkdir -p /logs

# Copy the built binary and CSV file from the builder stage
COPY --from=builder /app/process_csv /app/
COPY --from=builder /app/Prod1.csv /app/

# Command to run the application
CMD ["go", "run", "."]
