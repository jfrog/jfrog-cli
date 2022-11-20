# Wrap jfrog-testing-infra in a contaner for the docker tests.
# We run a container with Artifactory since docker tests may need to spin up a contaner to build images inside.

FROM golang:buster
RUN go install github.com/jfrog/jfrog-testing-infra/local-rt-setup@latest
ENV JFROG_HOME=/jfrog_home
WORKDIR /jfrog_home
EXPOSE 8082
EXPOSE 8081
CMD ["sh","-c","local-rt-setup; sleep infinity"]