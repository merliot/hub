# syntax=docker/dockerfile:1

FROM ghcr.io/merliot/device:main

WORKDIR /app
COPY . .

RUN go generate ./...
RUN go build -o /hub ./cmd
 
EXPOSE 8000

CMD ["/hub"]
