#!/bin/bash
./buildx create --name mybuilder
./buildx use mybuilder
./buildx inspect --bootstrap