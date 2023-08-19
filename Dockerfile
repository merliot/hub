# syntax=docker/dockerfile:1

FROM golang:1.20

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -tags wss -o /hub ./cmd/hub

EXPOSE 8000

# RUN wget https://github.com/tinygo-org/tinygo/releases/download/v0.28.1/tinygo_0.28.1_amd64.deb
# RUN dpkg -i tinygo_0.28.1_amd64.deb

CMD ["/hub"]
