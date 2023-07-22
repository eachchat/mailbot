FROM golang:1.20.0-alpine as builder

WORKDIR /app/github.com

COPY . .

# uncomment it in China
#RUN sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories
RUN apk add --no-cache gcc musl-dev git
# uncomment it in China
#RUN go env -w GOPROXY="https://goproxy.cn,direct"
RUN go mod tidy
RUN go mod vendor 
RUN CGO_ENABLED=1 go build -o main cmd/main.go
RUN pwd && ls -lah

FROM alpine:latest

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
WORKDIR /app

COPY --from=builder /app/github.com/main .

CMD [ "/app/main"]
