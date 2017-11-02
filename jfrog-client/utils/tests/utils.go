package tests

import (
	"net"
	"net/http"
	"os/exec"
	"os"
	"bufio"
	"strings"
	"regexp"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
)

type HttpServerHandlers map[string]func(w http.ResponseWriter, r *http.Request)

func StartHttpServer(handlers HttpServerHandlers) (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	go func() {
		httpMux := http.NewServeMux()
		for k, v := range handlers {
			httpMux.HandleFunc(k, v)
		}
		err = http.Serve(listener, httpMux)
		if err != nil {
			panic(err)
		}
	}()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func GetPackages() []string {
	cmd := exec.Command("go", "list", "../...")
	packages, _ := cmd.Output()

	scanner := bufio.NewScanner(strings.NewReader(string(packages)))
	var unitTests []string
	for scanner.Scan() {
		unitTests = append(unitTests, scanner.Text())
	}
	return unitTests
}


func ExcludeTestsPackage(packages []string, packageToExclude string) []string {
	var res []string
	for _, packageName := range packages {
		excludedTest, _ := regexp.MatchString(packageToExclude, packageName)
		if !excludedTest {
			res = append(res, packageName)
		}
	}
	return res
}

func RunTests(tests []string) {
	tests = append([]string{"test", "-v"}, tests...)
	cmd := exec.Command("go", tests...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}