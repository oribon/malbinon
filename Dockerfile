FROM golang:1.13.4-alpine as build-env
WORKDIR /malbinon
COPY . .
RUN go mod download
RUN go build -o malbinon

FROM alpine:latest
RUN mkdir /images /dirs
WORKDIR /malbinon
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=build-env /malbinon/ .
EXPOSE 443

CMD ["./malbinon"]
