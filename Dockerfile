# syntax=docker/dockerfile:1

FROM ghcr.io/merliot/base:v0.0.3

WORKDIR /app
COPY . .

RUN go version
RUN tinygo version
RUN go generate ./...
RUN go build -o /hub ./cmd/hub
 
EXPOSE 8000

CMD ["/hub"]
