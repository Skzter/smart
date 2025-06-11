FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum* ./

RUN go mod download

COPY . .

RUN go tool golangci-lint run

RUN go test ./...

# CGO_ENABLED=0 ensures static linking for scratch container
# -ldflags="-s -w" reduces binary size by stripping debug information
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/suproxy ./cmd/suproxy

FROM scratch

COPY --from=builder /app/bin/suproxy /bin/suproxy

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/bin/suproxy"]
