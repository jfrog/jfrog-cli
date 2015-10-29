package utils

import (
    "os"
)

func CheckError(err error) {
    if err != nil {
        panic(err)
    }
}

func Exit(msg string) {
    println(msg)
    os.Exit(1)
}

type BintrayDetails struct {
    Url string
    Org string
    User string
    Key string
}