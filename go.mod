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
	github.com/jfrog/jfrog-cli-core v1.11.2
	github.com/jfrog/jfrog-client-go v0.27.3
	github.com/jszwec/csvutil v1.4.0
	github.com/mholt/archiver v2.1.0+incompatible
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/vbauerster/mpb/v4 v4.7.0
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/crypto v0.0.0-20211202192323-5770296d904e
	gopkg.in/yaml.v2 v2.4.0
)

exclude golang.org/x/text v0.3.6

// replace github.com/jfrog/jfrog-client-go => github.com/jfrog/jfrog-client-go v1.0.1-0.20211005084545-98198805137b

replace github.com/jfrog/jfrog-cli-core => github.com/or-geva/jfrog-cli-core v0.0.0-20220224131718-76665be4ceaf

replace github.com/jfrog/gocmd => github.com/or-geva/gocmd v0.1.11-0.20220224131210-77b532c7a779

// replace github.com/jfrog/gofrog => github.com/jfrog/gofrog v1.0.6
