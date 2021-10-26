module github.com/jfrog/jfrog-cli

go 1.14

require (
	github.com/buger/jsonparser v1.1.1
	github.com/codegangsta/cli v1.20.0
	github.com/frankban/quicktest v1.13.1 // indirect
	github.com/go-git/go-git/v5 v5.4.2
	github.com/gookit/color v1.4.2
	github.com/jfrog/gofrog v1.0.7
	github.com/jfrog/jfrog-cli-core/v2 v2.4.1
	github.com/jfrog/jfrog-client-go v1.5.1
	github.com/jszwec/csvutil v1.4.0
	github.com/mholt/archiver v2.1.0+incompatible
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/vbauerster/mpb/v4 v4.7.0
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/jfrog/jfrog-client-go => github.com/sverdlov93/jfrog-client-go v1.0.2-0.20211025153925-b40d341deb7e

replace github.com/jfrog/jfrog-cli-core/v2 => github.com/sverdlov93/jfrog-cli-core/v2 v2.0.2-0.20211026110932-7195306ff262

// replace github.com/jfrog/gofrog => github.com/jfrog/gofrog v1.0.6
