module github.com/jfrog/jfrog-cli

go 1.14

require (
	github.com/buger/jsonparser v0.0.0-20180910192245-6acdf747ae99
	github.com/codegangsta/cli v1.20.0
	github.com/jfrog/gocmd v0.1.16
	github.com/jfrog/gofrog v1.0.6
	github.com/jfrog/jfrog-cli-core v1.0.1
	github.com/jfrog/jfrog-client-go v0.14.2
	github.com/mholt/archiver v2.1.0+incompatible
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.4.0
	github.com/vbauerster/mpb/v4 v4.7.0
	golang.org/x/crypto v0.0.0-20190510104115-cbcb75029529
	gopkg.in/yaml.v2 v2.2.2
)

// replace github.com/jfrog/jfrog-client-go => github.com/jfrog/jfrog-client-go dev

// replace github.com/jfrog/jfrog-cli-core => github.com/jfrog/jfrog-cli-core dev

// replace github.com/jfrog/gocmd => github.com/jfrog/gocmd v0.1.16

// replace github.com/jfrog/gofrog => github.com/jfrog/gofrog v1.0.6
