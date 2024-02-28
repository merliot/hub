# syntax=docker/dockerfile:1

FROM ghcr.io/merliot/device:main

WORKDIR /app
COPY . .

RUN go build -o /hub ./cmd
RUN /hub -uf2
 
EXPOSE 8000

CMD ["/hub"]
