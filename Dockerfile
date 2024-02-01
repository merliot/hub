# syntax=docker/dockerfile:1

FROM ghcr.io/merliot/device/device-base:latest

WORKDIR /app/hub

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

ARG SCHEME=wss

RUN go work use .
RUN go build -tags $SCHEME -o /hub ./cmd/

RUN go run ../device/cmd/uf2-builder -target nano-rp2040 -model garage
RUN go run ../device/cmd/uf2-builder -target wioterminal -model garage
RUN go run ../device/cmd/uf2-builder -target nano-rp2040 -model relays
RUN go run ../device/cmd/uf2-builder -target wioterminal -model relays

EXPOSE 8000

CMD ["/hub"]
