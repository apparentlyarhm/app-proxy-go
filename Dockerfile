# ---- Stage 1: The Builder ----
# big full image for the compilation
FROM golang:1.25-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the Go application..
# CGO_ENABLED=0: creates a statically linked binary, crucial for running in minimal images.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/main.go


# ---- Stage 2: The Final Image ----
# base image with shell
FROM alpine:latest

# It's good practice to run as a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

WORKDIR /home/appuser

# Copy ONLY the compiled binary from the builder stage.
COPY --from=builder /app/main .

EXPOSE 8080
CMD ["./main"]