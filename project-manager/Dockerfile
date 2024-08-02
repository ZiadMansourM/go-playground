# Build stage
FROM --platform=$BUILDPLATFORM golang:1.22.5 AS builder

WORKDIR /app

COPY go.mod go.sum /app/

RUN go mod download

COPY . .

ARG TARGETOS TARGETARCH

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o main

# Final stage
FROM --platform=$TARGETPLATFORM scratch
WORKDIR /app
COPY --from=builder /app/ /app/

EXPOSE 8080

CMD ["/app/main"]