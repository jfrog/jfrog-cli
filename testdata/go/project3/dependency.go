package dependency

import (
	"fmt"
	"github.com/pkg/errors"
)

func PrintHello() error {
	fmt.Println("Hello")
	return errors.New("abc")
}
