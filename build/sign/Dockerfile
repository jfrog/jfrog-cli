# This Dockerfile sets up a container, which is used to sign the Windows binary of JFrog CLI.

FROM ubuntu:16.04
RUN echo "deb http://cz.archive.ubuntu.com/ubuntu xenial main universe" >> /etc/apt/sources.list
RUN apt -y update && apt install -y curl build-essential libssl-dev libcurl4-gnutls-dev autoconf osslsigncode
ADD . /workspace