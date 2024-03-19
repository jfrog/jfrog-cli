#!/bin/bash
# We pin v0.12.2 since buildkit v0.12.3 is getting a 401 unauthorized error when using a private HTTP insecure registry.
./buildx create --name mybuilder --driver-opt "network=host" --driver-opt image=moby/buildkit:v0.12.2 --config buildkitd.toml --use
./buildx inspect --bootstrap