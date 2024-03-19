#!/bin/bash

echo -e "[jfrog-cli]
name=jfrog-cli
baseurl=https://releases.jfrog.io/artifactory/jfrog-rpms
enabled=1
gpgcheck=0" | sudo tee /etc/yum.repos.d/jfrog-cli.repo >/dev/null
yum install -y jfrog-cli-v2;
jf intro;
