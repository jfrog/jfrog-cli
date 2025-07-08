module github.com/jfrog/jfrog-cli

go 1.23.9

replace (
	// Should not be updated to 0.2.6 due to a bug (https://github.com/jfrog/jfrog-cli-core/pull/372)
	github.com/c-bata/go-prompt => github.com/c-bata/go-prompt v0.2.5

	// Should not be updated to 0.2.0-beta.2 due to a bug (https://github.com/jfrog/jfrog-cli-core/pull/372)
	github.com/pkg/term => github.com/pkg/term v1.1.0
)

require (
	github.com/agnivade/levenshtein v1.2.1
	github.com/buger/jsonparser v1.1.1
	github.com/docker/docker v27.5.1+incompatible
	github.com/gocarina/gocsv v0.0.0-20240520201108-78e41c74b4b1
	github.com/jfrog/archiver/v3 v3.6.1
	github.com/jfrog/build-info-go v1.10.14
	github.com/jfrog/gofrog v1.7.6
	github.com/jfrog/jfrog-cli-artifactory v0.3.2
	github.com/jfrog/jfrog-cli-core/v2 v2.59.0
	github.com/jfrog/jfrog-cli-platform-services v1.9.0
	github.com/jfrog/jfrog-cli-security v1.19.0
	github.com/jfrog/jfrog-client-go v1.54.1
	github.com/jszwec/csvutil v1.10.0
	github.com/manifoldco/promptui v0.9.0
	github.com/stretchr/testify v1.10.0
	github.com/testcontainers/testcontainers-go v0.35.0
	github.com/urfave/cli v1.22.16
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/exp v0.0.0-20250305212735-054e65f0b394
	gopkg.in/yaml.v2 v2.4.0
)

require (
	dario.cat/mergo v1.0.1 // indirect
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20240806141605-e8a1dd7889d6 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/BurntSushi/toml v1.4.0 // indirect
	github.com/CycloneDX/cyclonedx-go v0.9.2 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.1.6 // indirect
	github.com/VividCortex/ewma v1.2.0 // indirect
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/apache/camel-k/v2 v2.5.0 // indirect
	github.com/beevik/etree v1.4.0 // indirect
	github.com/c-bata/go-prompt v0.2.6 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/chzyer/readline v1.5.1 // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/cpuguy83/dockercfg v0.3.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.6 // indirect
	github.com/cyphar/filepath-securejoin v0.4.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dsnet/compress v0.0.2-0.20230904184137-39efe44ab707 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/forPelevin/gomoji v1.3.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gfleury/go-bitbucket-v1 v0.0.0-20230825095122-9bc1711434ab // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.6.2 // indirect
	github.com/go-git/go-git/v5 v5.14.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-logr/zapr v1.3.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-github/v56 v56.0.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gookit/color v1.5.4 // indirect
	github.com/grokify/mogo v0.64.12 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jedib0t/go-pretty/v6 v6.6.5 // indirect
	github.com/jfrog/froggit-go v1.17.0 // indirect
	github.com/jfrog/go-mockhttp v0.3.1 // indirect
	github.com/jfrog/jfrog-apps-config v1.0.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/klauspost/pgzip v1.2.6 // indirect
	github.com/ktrysmt/go-bitbucket v0.9.80 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magiconair/properties v1.8.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mattn/go-tty v0.0.5 // indirect
	github.com/microsoft/azure-devops-go-api/azuredevops/v7 v7.1.0 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/moby/sys/user v0.1.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/nwaples/rardecode v1.1.3 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/owenrumney/go-sarif/v2 v2.3.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pjbgf/sha1cd v0.3.2 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/term v1.2.0-beta.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/secure-systems-lab/go-securesystemslib v0.8.0 // indirect
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/shirou/gopsutil/v3 v3.23.12 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/skeema/knownhosts v1.3.1 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.19.0 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/ulikunitz/xz v0.5.12 // indirect
	github.com/vbauerster/mpb/v8 v8.9.1 // indirect
	github.com/virtuald/go-ordered-json v0.0.0-20170621173500-b18e6e673d74 // indirect
	github.com/xanzy/go-gitlab v0.110.0 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.49.0 // indirect
	go.opentelemetry.io/otel v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.34.0 // indirect
	go.opentelemetry.io/otel/metric v1.34.0 // indirect
	go.opentelemetry.io/otel/sdk v1.34.0 // indirect
	go.opentelemetry.io/otel/trace v1.34.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/oauth2 v0.27.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/term v0.30.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/time v0.7.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.29.7 // indirect
	k8s.io/apimachinery v0.29.7 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/utils v0.0.0-20240310230437-4693a0247e57 // indirect
	sigs.k8s.io/controller-runtime v0.17.5 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

// replace github.com/jfrog/build-info-go => github.com/jfrog/build-info-go v1.8.9-0.20250529104758-6d769a684388

replace github.com/jfrog/jfrog-cli-artifactory => github.com/jfrog/jfrog-cli-artifactory v0.3.3-0.20250708062743-ea3521f9b71a

replace github.com/jfrog/jfrog-cli-core/v2 => github.com/jfrog/jfrog-cli-core/v2 v2.31.1-0.20250629143350-c43614bbffcd

replace github.com/jfrog/jfrog-client-go => github.com/jfrog/jfrog-client-go v1.28.1-0.20250629142537-bb24db402fe1

//replace github.com/jfrog/jfrog-cli-security => github.com/jfrog/jfrog-cli-security v1.17.2-0.20250511132918-d9cc4cd50020
