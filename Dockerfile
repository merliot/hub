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

RUN wget "https://github.com/git-lfs/git-lfs/releases/download/v3.4.0/git-lfs-linux-amd64-v3.4.0.tar.gz"
RUN tar xvfz git-lfs-linux-amd64-v3.4.0.tar.gz
RUN git-lfs-3.4.0/install.sh

RUN git clone https://github.com/merliot/hub.git
RUN dpkg -i hub/release.deb

CMD ["/hub"]
