# ==============================================================================
# STAGE 1: Build Stage (Compiling the Go Code)
# ==============================================================================
FROM golang:alpine AS builder

WORKDIR /app

# Copy the module files first and download dependencies.
# This caches the downloads so we don't redownload every time we change code.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . .

# We use a build argument to dynamically decide WHICH microservice to compile.
# This single Dockerfile handles the API, the Router, and all Workers!
ARG APP_NAME
RUN if [ -z "$APP_NAME" ]; then echo "Build argument APP_NAME is required" && exit 1; fi

# Build the specific binary into a static executable.
# CGO_ENABLED=0 ensures it doesn't rely on external C libraries, making it ultra-portable.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /go/bin/microservice ./cmd/${APP_NAME}/main.go

# ==============================================================================
# STAGE 2: Final Minimal Runtime Image
# ==============================================================================
# Instead of using the heavy Go image to run the code, we use a tiny Alpine Linux image.
FROM alpine:latest

# Install CA Certificates in case we need to make outbound HTTPS calls (e.g. AWS SES, Twilio)
RUN apk --no-cache add ca-certificates

WORKDIR /app/

# Copy the compiled binary from the builder stage
COPY --from=builder /go/bin/microservice .

# Expose port 8080 (Required for the API Gateway; ignored safely by background workers)
EXPOSE 8080

# Execute the specific microservice
CMD ["./microservice"]
