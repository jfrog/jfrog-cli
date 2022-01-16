#!/bin/bash

if [ $# -eq 0 ]
  then
	exe_name="jf"
  else
	exe_name="$1"
fi

if CGO_ENABLED=0 go build -o $exe_name -ldflags '-w -extldflags "-static"' main.go
  then
  echo "The $exe_name executable was successfully created."
fi