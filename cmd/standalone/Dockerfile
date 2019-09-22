FROM golang:1.13
RUN mkdir /app
ADD . /app/
WORKDIR /app
CMD ["go", "run", "cmd/standalone/main.go"]
