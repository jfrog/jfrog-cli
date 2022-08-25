ARG repo_name_21
# Remove ${repo_name_21} to pull from Docker Hub.
FROM ${repo_name_21}/jfrog-docker/golang:1.18.5-alpine as builder
ARG image_name=jfrog-cli
ARG cli_executable_name
WORKDIR /${image_name}
COPY . /${image_name}
RUN apk add --update git && sh build/build.sh ${cli_executable_name}
# Remove ${repo_name_21} to pull from Docker Hub.
FROM ${repo_name_21}/jfrog-docker/alpine:3.12.0
ARG image_name=jfrog-cli
ARG cli_executable_name
ENV CI true
RUN apk add --no-cache bash tzdata ca-certificates curl
COPY --from=builder /${image_name}/${cli_executable_name} /usr/local/bin/${cli_executable_name}
RUN chmod +x /usr/local/bin/${cli_executable_name}
