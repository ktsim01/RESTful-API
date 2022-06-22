# Start from golang base image
FROM golang:alpine
# ENV GO111MODULE=on

# Set the current working directory inside the container 
WORKDIR /app

# Copy go mod and sum files 
# COPY go.mod go.sum ./


# Copy the source from the current directory to the working Directory inside the container 
COPY . .

# Download all dependencies. Dependencies will be cached if the go.mod and the go.sum files are not changed 
RUN go mod download

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Expose port 8080 to the outside world
EXPOSE 8080

#Command to run the executable
CMD ["./main"]

