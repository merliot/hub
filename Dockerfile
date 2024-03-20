# syntax=docker/dockerfile:1

FROM ghcr.io/merliot/base:v0.0.1

WORKDIR /app
COPY . .

RUN go generate ./cmd/gen-ver/
RUN go build -o /hub ./cmd
RUN go generate ./cmd/gen-uf2/
 
EXPOSE 8000

CMD ["/hub"]
