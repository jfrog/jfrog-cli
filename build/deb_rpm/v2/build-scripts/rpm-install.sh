#!/bin/bash

{
  echo "[jfrog-cli]"
  echo "name=jfrog-cli"
  echo "baseurl=https://releases.jfrog.io/artifactory/jfrog-rpms"
  echo "enabled=1"
  echo "gpgcheck=0"
} >> jfrog-cli.repo
sudo mv jfrog-cli.repo /etc/yum.repos.d/;
yum install -y jfrog-cli-v2;
jf intro;
