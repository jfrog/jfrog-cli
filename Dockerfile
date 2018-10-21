FROM golang:1.11-alpine as builder
WORKDIR /go/src/github.com/jfrog/jfrog-cli-go
COPY . /go/src/github.com/jfrog/jfrog-cli-go
RUN CGO_ENABLED=0 GOOS=linux go build github.com/jfrog/jfrog-cli-go/jfrog-cli/jfrog

FROM alpine:3.7
RUN apk add --no-cache bash tzdata ca-certificates
COPY --from=builder /go/src/github.com/jfrog/jfrog-cli-go/jfrog /usr/local/bin/jfrog
RUN chmod +x /usr/local/bin/jfrog

ENTRYPOINT [ "/usr/local/bin/jfrog" ]
CMD ["--help"]
