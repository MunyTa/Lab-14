FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN go build -o /collector ./cmd/collector

FROM alpine:3.21

WORKDIR /app
COPY --from=build /collector /usr/local/bin/collector
ENTRYPOINT ["collector"]

