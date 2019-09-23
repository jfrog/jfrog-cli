FROM golang:1.12-alpine as builder

# Add NodeJS/NPM to support npm-publish
#   jfrog rt npm-publish my-local-npm-repo --build-name=my-build-name --build-number=1
# https://www.jfrog.com/confluence/display/CLI/CLI+for+JFrog+Artifactory#CLIforJFrogArtifactory-BuildingNpmPackages
# https://github.com/jfrog/jfrog-cli/issues/348
RUN apk add --update nodejs && \
    rm -rf /var/cache/apk/* && \
    npm i npm -g

WORKDIR /jfrog-cli-go
COPY . /jfrog-cli-go
RUN apk add --update git && sh build.sh
FROM alpine:3.7
RUN apk add --no-cache bash tzdata ca-certificates
COPY --from=builder /jfrog-cli-go/jfrog /usr/local/bin/jfrog
RUN chmod +x /usr/local/bin/jfrog
