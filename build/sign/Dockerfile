# This Dockerfile sets up a container, which is used to sign the Windows binary of JFrog CLI.

FROM ubuntu:16.04
RUN echo "deb http://cz.archive.ubuntu.com/ubuntu xenial main universe" >> /etc/apt/sources.list
RUN apt-get -y update \
  && apt-get install -y curl \
  && apt-get install -y build-essential \ 
  && apt-get install -y libssl-dev \
  && apt-get install -y libcurl4-gnutls-dev \
  && apt-get install -y autoconf \
  && apt-get install -y osslsigncode

ADD . /workspace