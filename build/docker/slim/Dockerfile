ARG image_name=jfrog-cli
# In order to build this image localy, remove ${REPO_NAME_21}
FROM ${REPO_NAME_21}/jfrog-docker/golang:1.17.2-alpine as builder
WORKDIR /${image_name}
COPY . /${image_name}
RUN apk add --update git && sh build/build.sh
# In order to build this image localy, remove ${REPO_NAME_21}
FROM ${REPO_NAME_21}/jfrog-docker/alpine:3.12.0
ENV CI true
RUN apk add --no-cache bash tzdata ca-certificates curl
COPY --from=builder /${image_name}/jfrog /usr/local/bin/jfrog
RUN chmod +x /usr/local/bin/jfrog
