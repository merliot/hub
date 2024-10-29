# syntax=docker/dockerfile:1

FROM ghcr.io/merliot/base

WORKDIR /app
COPY . .

RUN go generate ./...
RUN go build -o /hub ./cmd/hub
 
EXPOSE 8000

CMD ["/hub"]
