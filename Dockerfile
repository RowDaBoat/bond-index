FROM golang:1.21-alpine AS builder

WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o yarr

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /src/yarr /yarr
RUN mkdir -p /data

EXPOSE 80

CMD ["/yarr"]
