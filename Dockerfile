FROM golang:1.15 as builder

COPY . /go/app
WORKDIR /go/app

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/userservice

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root
COPY --from=builder /go/app/app .
CMD ["./app"]

