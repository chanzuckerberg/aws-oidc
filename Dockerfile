# First stage: build the executable
FROM golang:1.17 AS builder

# Enable Go modules
ENV GO111MODULE=on CGO_ENABLED=0 GOOS=linux

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container
COPY cmd cmd
COPY go.mod go.sum main.go ./
COPY pkg pkg

# Build the Go app
RUN go build -o aws-oidc .

# Final stage: the running container
FROM alpine:latest AS final

# Install SSL root certificates
RUN apk update && apk --no-cache add ca-certificates curl

COPY --from=builder /app/aws-oidc /bin/aws-oidc

ADD https://github.com/segmentio/chamber/releases/download/v2.7.5/chamber-v2.7.5-linux-amd64 /bin/chamber
RUN chmod +x /bin/chamber


CMD ["aws-oidc"]
