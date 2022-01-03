FROM golang:1.17-alpine as builder
WORKDIR /app

COPY go.* /app/
RUN go mod download
COPY . /app/

RUN CGO_ENABLED=0 GOOS=linux go build -a -o registry ./cmd


FROM alpine:3.9 

WORKDIR /root/

COPY --from=builder /app/ /root/
COPY --from=builder /app/config.yaml /root/

EXPOSE 8000

ENTRYPOINT ["./registry"]


