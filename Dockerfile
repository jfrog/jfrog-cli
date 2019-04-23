FROM golang:1.11-alpine as builder
WORKDIR /jfrog-cli-go
COPY . /jfrog-cli-go
RUN apk add --update git && \
    CGO_ENABLED=0 GOOS=linux go build /jfrog-cli-go/jfrog

FROM alpine:3.7
RUN apk add --no-cache bash tzdata ca-certificates
COPY --from=builder /jfrog-cli-go/jfrog /usr/local/bin/jfrog
RUN chmod +x /usr/local/bin/jfrog
