#!/bin/bash
./buildx create --name mybuilder --driver-opt "network=host" --config buildkitd.toml --use
./buildx inspect --bootstrap