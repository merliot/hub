# syntax=docker/dockerfile:1

# Debian GNU/Linux 12 (bookworm)
FROM golang:1.24.1

RUN wget https://github.com/tinygo-org/tinygo/releases/download/v0.36.0/tinygo_0.36.0_amd64.deb
RUN dpkg -i tinygo_0.36.0_amd64.deb

RUN apt-get update
RUN apt-get install vim tree bc ffmpeg -y

WORKDIR /app
COPY . .

# Generate UF2 base images and build the hub
RUN go generate ./...
RUN go build -o /hub ./cmd

# Expose the port for /hub
EXPOSE 8000

# CMD provides the default argument to the entrypoint
CMD ["/hub"]
