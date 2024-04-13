FROM golang:1.20

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY app/go.mod app/go.sum ./
RUN go mod download

# Copy the source code
COPY app/*.go ./

# Build the application
RUN go build -o bin .

ENTRYPOINT ["/app/bin"]

# Expose the port the application will run on
EXPOSE 8080

# Run the application
