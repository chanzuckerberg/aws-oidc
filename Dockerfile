FROM 533267185808.dkr.ecr.us-west-2.amazonaws.com/docker.io/central/library/golang:1.23-alpine AS builder

WORKDIR /app

COPY cmd cmd
COPY go.mod go.sum main.go ./
COPY pkg pkg

ARG PLATFORM=arm64
ARG RELEASE_VERSION
ARG GITHUB_SHA
RUN --mount=type=cache,mode=0755,target=/go/pkg/mod CGO_CFLAGS="-D_LARGEFILE64_SOURCE" CGO_ENABLED=0 GO111MODULE=on GOOS=linux GOARCH=${PLATFORM} go build -o aws-oidc

FROM 533267185808.dkr.ecr.us-west-2.amazonaws.com/docker.io/central/library/golang:1.23-alpine AS prod

RUN apk update && apk --no-cache add ca-certificates curl

COPY --from=builder /app/aws-oidc /bin/aws-oidc
COPY ./entrypoint.sh ./entrypoint.sh

CMD ["serve-config"]
ENTRYPOINT ["./entrypoint.sh"]