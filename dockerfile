FROM golang:alpine

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go build -o app .

# Move to /bin directory as the place for resulting binary folder
WORKDIR /bin

# Move env files to dist directory
COPY .env .
COPY .env.development .

# Copy binary from build to main folder
RUN cp /build/app .

# Export necessary port
EXPOSE 4567

# Command to run when starting the container
CMD ["/bin/app"]
