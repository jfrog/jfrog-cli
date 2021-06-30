module github.com/jfrog/jfrog-cli

go 1.14

require (
	github.com/buger/jsonparser v0.0.0-20180910192245-6acdf747ae99
	github.com/codegangsta/cli v1.20.0
	github.com/frankban/quicktest v1.11.3 // indirect
	github.com/go-git/go-git/v5 v5.4.2
	github.com/gookit/color v1.4.2
	github.com/jfrog/gocmd v0.3.1
	github.com/jfrog/gofrog v1.0.6
	github.com/jfrog/jfrog-cli-core v1.8.0
	github.com/jfrog/jfrog-client-go v0.25.0
	github.com/jszwec/csvutil v1.4.0
	github.com/mholt/archiver v2.1.0+incompatible
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/vbauerster/mpb/v4 v4.7.0
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/jfrog/jfrog-client-go => github.com/jfrog/jfrog-client-go v0.25.1-0.20210630114617-2eccf3b6f075

//replace github.com/jfrog/jfrog-cli-core => github.com/jfrog/jfrog-cli-core v1.8.1-0.20210630094549-8ae890e61c9d
replace github.com/jfrog/jfrog-cli-core => github.com/jfrog/jfrog-cli-core v1.8.1-0.20210630141146-7fc26625e09f

//replace github.com/jfrog/gocmd => github.com/jfrog/gocmd v0.3.2-0.20210630094119-287a232e84c7
replace github.com/jfrog/gocmd => github.com/jfrog/gocmd v0.3.2-0.20210630094119-287a232e84c7

// replace github.com/jfrog/gofrog => github.com/jfrog/gofrog v1.0.6
