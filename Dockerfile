############################
# STAGE 1
############################
FROM golang:alpine AS builder

RUN apk --update add --no-cache ca-certificates openssl git tzdata && \
update-ca-certificates

RUN mkdir -p /build
WORKDIR /build

# Copy all files over
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o spot-interruption-exporter
############################
# STAGE 2
############################
FROM scratch

# Copy all certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copies the built binary over from the previous stage
COPY --from=builder /build/spot-interruption-exporter /go/bin/spot-interruption-exporter

# Run the binary
ENTRYPOINT ["/go/bin/spot-interruption-exporter"]
