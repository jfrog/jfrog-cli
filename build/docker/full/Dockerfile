ARG cli_executable_name
ARG repo_name_21
ARG image_name=jfrog-cli-full
# Remove ${repo_name_21} to pull from Docker Hub.
FROM ${repo_name_21}/jfrog-docker/golang:1.17.2 as builder
WORKDIR /${image_name}
COPY . /${image_name}
RUN sh build/build.sh ${cli_executable_name}
FROM releases-docker.jfrog.io/jfrog-ecosystem-integration-env:latest
ENV CI true
COPY --from=builder /${image_name}/${cli_executable_name} /usr/local/bin/${cli_executable_name}
RUN chmod +x /usr/local/bin/${cli_executable_name}
