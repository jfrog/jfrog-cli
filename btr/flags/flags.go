package flags

import (
    "github.com/codegangsta/cli"
)

type Flags interface {

    Get(args interface{}) []cli.Flag
}
