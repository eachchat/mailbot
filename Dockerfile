FROM golang:1.20.0-alpine as builder

WORKDIR /app/matrixemailbridge

COPY ./* ./

RUN sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories
RUN apk add --no-cache gcc musl-dev git
RUN go env -w GOPROXY="https://goproxy.cn,direct"
RUN go mod tidy
RUN go mod vendor 
RUN CGO_ENABLED=1
RUN go build -o main cmd/main.go
RUN pwd && ls -lah

FROM alpine:latest

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
WORKDIR /app

COPY --from=builder /app/matrixemailbridge/main .

RUN mkdir /app/data/
RUN ls -lath

ENV BRIDGE_DATA_PATH="/app/data/"

CMD [ "/app/main"]
