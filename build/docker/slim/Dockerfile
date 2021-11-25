ARG cliExecutableName
ARG image_name=jfrog-cli
# Remove ${REPO_NAME_21} to pull from Docker Hub.
FROM ${REPO_NAME_21}/jfrog-docker/golang:1.17.2-alpine as builder
WORKDIR /${image_name}
COPY . /${image_name}
RUN apk add --update git && sh build/build.sh ${cliExecutableName}
# Remove ${REPO_NAME_21} to pull from Docker Hub.
FROM ${REPO_NAME_21}/jfrog-docker/alpine:3.12.0
ENV CI true
RUN apk add --no-cache bash tzdata ca-certificates curl
COPY --from=builder /${image_name}/${cliExecutableName} /usr/local/bin/${cliExecutableName}
RUN chmod +x /usr/local/bin/${cliExecutableName}
