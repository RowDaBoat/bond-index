FROM golang:1.23-alpine AS builder

WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o yarr

FROM alpine:latest

ENV ORD_URL=""
COPY --from=builder /src/yarr /yarr
RUN mkdir -p /data
EXPOSE 80

CMD ["/yarr", "--ord-url", "$ORD_URL"]
