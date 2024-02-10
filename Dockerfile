# syntax=docker/dockerfile:1

FROM ghcr.io/merliot/device/device-base:latest

WORKDIR /app
RUN git clone https://github.com/merliot/device.git
RUN go work use device

RUN git clone https://github.com/merliot/hub.git
RUN go work use hub

WORKDIR /app/hub

ARG SCHEME=wss

RUN go build -tags $SCHEME -o /hub ./cmd/

RUN go run ../device/cmd/uf2-builder -target nano-rp2040 -model garage
RUN go run ../device/cmd/uf2-builder -target wioterminal -model garage
RUN go run ../device/cmd/uf2-builder -target nano-rp2040 -model relays
RUN go run ../device/cmd/uf2-builder -target wioterminal -model relays
RUN go run ../device/cmd/uf2-builder -target nano-rp2040 -model ps30m

EXPOSE 8000

CMD ["/hub"]
