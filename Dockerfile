# Use the official golang image as the base image
FROM golang:1.19

# Set the working directory inside the container
WORKDIR /app

# Copy the local package files to the container's workspace
COPY . .

# Build the Go application
RUN go build -o KVDatastore .


# Expose the necessary ports
EXPOSE 4040
EXPOSE 4041
EXPOSE 3000
EXPOSE 3001

# Command to run the application
CMD ["./KVDatastore"]