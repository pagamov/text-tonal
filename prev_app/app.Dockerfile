FROM golang:alpine
WORKDIR /app
COPY app/main.go .
COPY app/go.mod .
COPY app/go.sum .
RUN go build -o main main.go

ARG DEFAULT_PORT=8080
ENV PORT $DEFAULT_PORT

EXPOSE $PORT

CMD ["./main"]


# FROM golang:alpine as builder

# WORKDIR /app
# COPY app/go.mod .
# COPY app/go.sum .
# RUN go mod download
# COPY app/. .
# ENV GOCACHE=/root/.cache/go-build
# RUN --mount=type=cache,target="/root/.cache/go-build" go build -o app

# FROM golang:alpine
# RUN mkdir /app
# WORKDIR /app
# COPY --from=builder /app/app .
# ENTRYPOINT ["./app"]

