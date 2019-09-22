FROM golang:1.13.0-stretch as builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main cmd/services/post_chirp/main.go
FROM scratch
COPY --from=builder /build/main /app/
WORKDIR /app
CMD ["./main"]
