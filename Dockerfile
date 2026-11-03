# Use the official Golang image to run the Go program
FROM golang:1.27

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source code into the container
COPY . .

# Download all the dependencies
RUN go mod download github.com/ovh/go-ovh

# Command to run the Go program
CMD ["go", "run", "main.go"]