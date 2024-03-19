#!/bin/bash

docker rm --force -v "artifactory"
docker network rm "test-network"