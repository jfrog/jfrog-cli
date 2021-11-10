module github.com/jfrog/jfrog-cli

go 1.14

require (
	github.com/buger/jsonparser v1.1.1
	github.com/codegangsta/cli v1.20.0
	github.com/frankban/quicktest v1.13.1 // indirect
	github.com/go-git/go-git/v5 v5.4.2
	github.com/gookit/color v1.4.2
	github.com/jfrog/gofrog v1.0.7
	github.com/jfrog/jfrog-cli-core/v2 v2.4.2
	github.com/jfrog/jfrog-client-go v1.5.2
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

replace (
	github.com/jfrog/gocmd => github.com/yahavi/gocmd v0.1.13-0.20211109135307-6b3e0aca9965
	github.com/jfrog/jfrog-cli-core/v2 => github.com/yahavi/jfrog-cli-core/v2 v2.0.0-20211110105913-e8e5b8e03bfc
	github.com/jfrog/jfrog-client-go => github.com/jfrog/jfrog-client-go v1.5.3-0.20211108203834-9e234c549753
)
