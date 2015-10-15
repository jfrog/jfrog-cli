package flags

import (
    "github.com/codegangsta/cli"
)

const defaultApiUrl = "https://bintray.com/api/v1/"

type Flags interface {

    Get(args interface{}) []cli.Flag
}