ARG repo_name_21
# Remove ${repo_name_21} to pull from Docker Hub.
FROM ${repo_name_21}/jfrog-docker/golang:1.17.9 as builder
ARG image_name=jfrog-cli-full
WORKDIR /${image_name}
COPY . /${image_name}
RUN sh build/build.sh
FROM releases-docker.jfrog.io/jfrog-ecosystem-integration-env:latest
ARG image_name=jfrog-cli-full
ENV CI true
COPY --from=builder /${image_name}/jfrog /usr/local/bin/jfrog
RUN chmod +x /usr/local/bin/jfrog