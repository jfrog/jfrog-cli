package dependency

import (
	"fmt"
	"github.com/pkg/errors"
	"rsc.io/quote"
)

func PrintHello() error {
	fmt.Println(quote.Hello())
	return errors.New("abc")
}
