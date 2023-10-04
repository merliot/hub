# syntax=docker/dockerfile:1

FROM ghcr.io/merliot/hub/hub-base:latest

WORKDIR /app/hub

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN go work use .
RUN CGO_ENABLED=0 GOOS=linux go build -o /hub ./cmd/hub

EXPOSE 8000

CMD ["/hub"]
