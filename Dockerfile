FROM golang:1.23-alpine3.19 AS builder
RUN apk add --no-cache git

WORKDIR /go/src/github.com/buildwithgrove/path-auth-data-server
COPY . .
RUN apk add --no-cache make build-base
RUN go build -o /go/bin/path-auth-data-server .

FROM alpine:3.19
WORKDIR /app

COPY --from=builder /go/bin/path-auth-data-server ./

CMD ["./path-auth-data-server"]
