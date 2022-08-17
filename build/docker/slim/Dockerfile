ARG repo_name_21
# Remove ${repo_name_21} to pull from Docker Hub.
FROM ${repo_name_21}/jfrog-docker/golang:1.17.13-alpine as builder
ARG image_name=jfrog-cli
WORKDIR /${image_name}
COPY . /${image_name}
RUN apk add --update git && sh build/build.sh
# Remove ${repo_name_21} to pull from Docker Hub.
FROM ${repo_name_21}/jfrog-docker/alpine:3.12.0
ARG image_name=jfrog-cli
ENV CI true
RUN apk add --no-cache bash tzdata ca-certificates curl
COPY --from=builder /${image_name}/jfrog /usr/local/bin/jfrog
RUN chmod +x /usr/local/bin/jfrog
