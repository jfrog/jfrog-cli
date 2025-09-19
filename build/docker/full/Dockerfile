ARG repo_name_21
# Remove ${repo_name_21} to pull from Docker Hub.
FROM ${repo_name_21}/jfrog-docker/golang:1.24 as builder
ARG image_name=jfrog-cli-full
ARG cli_executable_name
WORKDIR /${image_name}
COPY . /${image_name}
RUN sh build/build.sh ${cli_executable_name}
FROM releases-docker.jfrog.io/jfrog-ecosystem-integration-env:latest
ARG image_name=jfrog-cli-full
ARG cli_executable_name
ENV CI true
COPY --chmod=755 --from=builder /${image_name}/${cli_executable_name} /usr/local/bin/${cli_executable_name}
# Install bash completion
RUN ${cli_executable_name} completion bash >> /etc/bash_completion