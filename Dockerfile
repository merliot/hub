# syntax=docker/dockerfile:1

FROM merliot/base:v0.0.3

WORKDIR /app
COPY . .

# Print Go and TinyGo versions
RUN go version
RUN tinygo version

# Generate UF2 base images and build the hub
#RUN go generate ./...
RUN go build -o /hub ./cmd

# Expose the port for /hub
EXPOSE 8000

# Copy the entrypoint script
#COPY entrypoint.sh /entrypoint.sh
#RUN chmod +x /entrypoint.sh

# Use the entrypoint script as the entry point
#ENTRYPOINT ["/entrypoint.sh"]

# CMD provides the default argument to the entrypoint
CMD ["/hub"]
