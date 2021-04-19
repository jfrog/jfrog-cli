module github.com/jfrog/jfrog-cli

go 1.14

require (
	github.com/buger/jsonparser v0.0.0-20180910192245-6acdf747ae99
	github.com/codegangsta/cli v1.20.0
	github.com/frankban/quicktest v1.11.3 // indirect
	github.com/go-git/go-git/v5 v5.2.0
	github.com/gookit/color v1.4.2
	github.com/jfrog/gocmd v0.2.0
	github.com/jfrog/gofrog v1.0.6
	github.com/jfrog/jfrog-cli-core v1.5.2
	github.com/jfrog/jfrog-client-go v0.21.2
	github.com/jszwec/csvutil v1.4.0
	github.com/mholt/archiver v2.1.0+incompatible
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	github.com/vbauerster/mpb/v4 v4.7.0
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/crypto v0.0.0-20201208171446-5f87f3452ae9
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/jfrog/jfrog-client-go => github.com/jfrog/jfrog-client-go v0.0.0-20210413100935-cc9831f340b3

replace github.com/jfrog/jfrog-cli-core => github.com/asafgabai/jfrog-cli-core v1.2.7-0.20210419071945-74e970d0421d

// replace github.com/jfrog/gocmd => github.com/jfrog/gocmd master

// replace github.com/jfrog/gofrog => github.com/jfrog/gofrog v1.0.6
