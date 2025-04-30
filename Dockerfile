# First stage: build the executable
FROM 533267185808.dkr.ecr.us-west-2.amazonaws.com/docker.io/central/library/golang:1.23-alpine AS builder

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
FROM 533267185808.dkr.ecr.us-west-2.amazonaws.com/docker.io/central/library/golang:1.23-alpine AS prod

# Install SSL root certificates
RUN apk update && apk --no-cache add ca-certificates curl

COPY --from=builder /app/aws-oidc /bin/aws-oidc

CMD ["aws-oidc"]
