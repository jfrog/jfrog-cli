module github.com/jfrog/jfrog-cli

go 1.14

require (
	github.com/buger/jsonparser v1.1.1
	github.com/codegangsta/cli v1.20.0
	github.com/frankban/quicktest v1.11.3 // indirect
	github.com/go-git/go-git/v5 v5.4.2
	github.com/gookit/color v1.4.2
	github.com/jfrog/gocmd v0.3.1
	github.com/jfrog/gofrog v1.0.6
	github.com/jfrog/jfrog-cli-core v1.11.4
	github.com/jfrog/jfrog-client-go v0.27.5
	github.com/jszwec/csvutil v1.4.0
	github.com/mholt/archiver v2.1.0+incompatible
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/vbauerster/mpb/v4 v4.7.0
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/crypto v0.0.0-20220314234659-1baeb1ce4c0b
	gopkg.in/yaml.v2 v2.4.0
)

exclude golang.org/x/text v0.3.6

// replace github.com/jfrog/jfrog-client-go => github.com/jfrog/jfrog-client-go v1.0.1-0.20220314173239-6c13b15c7673

// replace github.com/jfrog/jfrog-cli-core => github.com/jfrog/jfrog-cli-core v1.11.3-0.20220314180647-c7e7e3dbdb6c

replace github.com/jfrog/gocmd => github.com/jfrog/gocmd v0.3.2-0.20220314173619-0e00b67546d0

// replace github.com/jfrog/gofrog => github.com/jfrog/gofrog v1.0.6
