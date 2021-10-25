module github.com/jfrog/jfrog-cli

go 1.14

require (
	github.com/buger/jsonparser v1.1.1
	github.com/codegangsta/cli v1.20.0
	github.com/go-git/go-git/v5 v5.4.2
	github.com/gookit/color v1.4.2
	github.com/jfrog/build-info-go v0.0.0-20211020140610-2b15ac5444b5
	github.com/jfrog/gofrog v1.0.7
	github.com/jfrog/jfrog-cli-core/v2 v2.4.1
	github.com/jfrog/jfrog-client-go v1.5.1
	github.com/jszwec/csvutil v1.4.0
	github.com/kr/pretty v0.3.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/vbauerster/mpb/v4 v4.7.0
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/jfrog/jfrog-client-go => github.com/asafgabai/jfrog-client-go v0.18.1-0.20211025124419-de1bfcd0d18e

replace github.com/jfrog/jfrog-cli-core/v2 => github.com/asafgabai/jfrog-cli-core/v2 v2.0.0-20211025132501-7d92d42a19e5

replace github.com/jfrog/gocmd => github.com/asafgabai/gocmd v0.1.20-0.20211025124110-b76b3a6186df

//replace github.com/jfrog/gofrog => github.com/jfrog/gofrog v1.0.6

replace github.com/jfrog/build-info-go => github.com/asafgabai/build-info-go v0.0.0-20211025124318-baaa6130209e
