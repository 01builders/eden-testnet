FROM golang:1.25-alpine AS build-env

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o eden-testnet .

FROM alpine:3.18.3

RUN apk --no-cache add ca-certificates curl

WORKDIR /root

COPY --from=build-env /src/eden-testnet /usr/bin/eden-testnet
COPY ./entrypoint.sh /usr/bin/entrypoint.sh
RUN chmod +x /usr/bin/entrypoint.sh

ENTRYPOINT ["/usr/bin/entrypoint.sh"]
