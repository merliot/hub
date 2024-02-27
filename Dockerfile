# syntax=docker/dockerfile:1

FROM ghcr.io/merliot/device/device-base:latest

WORKDIR /app
COPY . .
RUN go work use .

RUN go build -o /hub ./cmd
RUN /hub -uf2
 
EXPOSE 8000

CMD ["/hub"]
