FROM golang:alpine
WORKDIR /pagamov
COPY main.go .
COPY go.mod .
COPY go.sum .
RUN go build -o main main.go

ARG DEFAULT_PORT=8080
ENV PORT $DEFAULT_PORT

EXPOSE $PORT

CMD ["./main"]

