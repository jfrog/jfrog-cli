# This Dockerfile sets up a container, which is used to sign the Windows binary of JFrog CLI.

FROM ubuntu:16.04
RUN apt-get -y update \
  && apt-get install -y curl \
  && apt-get install -y build-essential \ 
  && apt-get install -y libssl-dev \
  && apt-get install -y libcurl4-gnutls-dev \
  && apt-get install -y autoconf

RUN curl -s -L https://downloads.sourceforge.net/project/osslsigncode/osslsigncode/osslsigncode-1.7.1.tar.gz > osslsigncode.tar.gz
RUN tar xzf osslsigncode.tar.gz
RUN cd osslsigncode-1.7.1 && ./configure && make && make install

ADD . /workspace