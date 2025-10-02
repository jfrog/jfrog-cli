#!/bin/bash

awk '/require \(/,/\)/ {if ($1 != "require" && $1 != ")") print $0}' go.mod | \
grep 'github.com/jfrog' | \
grep -v '// indirect' | \
awk '{print $1}' | \
while read -r module; do
  echo "Upgrading $module..."
  go get "$module"
done

echo "Tidying up go.mod and go.sum..."
go mod tidy