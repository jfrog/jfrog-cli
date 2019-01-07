package projectmissingdependency

import (
	"fmt"
	"github.com/jfrog/dependency"
)

func execDependency() {
	dependency.PrintHello()
}

func Exec() {
	fmt.Println("Executing ")
	execDependency()
}