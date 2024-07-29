# Wrap jfrog-testing-infra in a contaner for the docker tests.
# We run a container with Artifactory since docker tests may need to spin up a contaner to build images inside.

FROM golang:buster
RUN go install github.com/jfrog/jfrog-testing-infra/local-rt-setup@latest
ENV JFROG_HOME=/jfrog_home
WORKDIR /jfrog_home
EXPOSE 8082
EXPOSE 8081
# Temporarily use version 7.84.17 due to an issue with the automatic token generation mechanism in 7.90
CMD ["sh","-c","local-rt-setup --rt-version 7.84.17; sleep infinity"]
