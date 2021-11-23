ARG cliExecutableName
ARG image_name=jfrog-cli-full
# Remove ${REPO_NAME_21} to pull from Docker Hub.
FROM ${REPO_NAME_21}/jfrog-docker/golang:1.17.2 as builder
WORKDIR /${image_name}
COPY . /${image_name}
RUN sh build/build.sh ${cliExecutableName}
FROM releases-docker.jfrog.io/jfrog-ecosystem-integration-env:latest
ENV CI true
COPY --from=builder /${image_name}/${cliExecutableName} /usr/local/bin/${cliExecutableName}
RUN chmod +x /usr/local/bin/${cliExecutableName}
