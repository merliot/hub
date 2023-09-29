# syntax=docker/dockerfile:1

FROM ghcr.io/merliot/tinygo-docker/tinygo-docker:latest

WORKDIR /app/hub

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN go work use .
RUN CGO_ENABLED=0 GOOS=linux go build -tags wss -o /hub ./cmd/hub

EXPOSE 8000

CMD ["/hub"]
