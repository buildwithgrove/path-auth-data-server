FROM golang:1.23-alpine3.19 AS builder
RUN apk add --no-cache git

WORKDIR /go/src/github.com/buildwithgrove/auth-server
COPY . .
RUN apk add --no-cache make build-base
RUN go build  -o /go/bin/path-auth-grpc-server ./main.go

FROM alpine:3.19
WORKDIR /app

COPY --from=builder /go/bin/path-auth-grpc-server ./

CMD ["./path-auth-grpc-server"]
