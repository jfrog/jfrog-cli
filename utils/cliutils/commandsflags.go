package cliutils

import (
	"github.com/codegangsta/cli"
	"sort"
)

const (
	// Base flags
	url         = "url"
	distUrl     = "dist-url"
	user        = "user"
	password    = "password"
	apikey      = "apikey"
	accessToken = "access-token"
	serverId    = "server-id"

	// Ssh flags
	sshKeyPath    = "ssh-key-path"
	sshPassPhrase = "ssh-passphrase"

	// Client certification flags
	clientCertPath    = "client-cert-path"
	clientCertKeyPath = "client-cert-key-path"
	insecureTls       = "insecure-tls"

	// Sort & limit flags
	sortBy    = "sort-by"
	sortOrder = "sort-order"
	limit     = "limit"
	offset    = "offset"

	// Spec flags
	spec     = "spec"
	specVars = "spec-vars"

	// Build info flags
	buildName   = "build-name"
	buildNumber = "build-number"
	module      = "module"

	// Generic commands flags
	excludePatterns = "exclude-patterns"
	exclusions      = "exclusions"
	recursive       = "recursive"
	build           = "build"
	includeDirs     = "include-dirs"
	props           = "props"
	excludeProps    = "exclude-props"
	failNoOp        = "fail-no-op"
	threads         = "threads"
	syncDeletes     = "sync-deletes"
	quite           = "quite"
	bundle          = "bundle"
	archiveEntries  = "archive-entries"

	// Config flags
	interactive = "interactive"
	encPassword = "enc-password"
)

var flagsMap = map[string]cli.Flag{
	url: cli.StringFlag{
		Name:  url,
		Usage: "[Optional] Artifactory URL.` `",
	},
	distUrl: cli.StringFlag{
		Name:  distUrl,
		Usage: "[Optional] Distribution URL.` `",
	},
	user: cli.StringFlag{
		Name:  user,
		Usage: "[Optional] Artifactory username.` `",
	},
	password: cli.StringFlag{
		Name:  password,
		Usage: "[Optional] Artifactory password.` `",
	},
	apikey: cli.StringFlag{
		Name:  apikey,
		Usage: "[Optional] Artifactory API key.` `",
	},
	accessToken: cli.StringFlag{
		Name:  accessToken,
		Usage: "[Optional] Artifactory access token.` `",
	},
	serverId: cli.StringFlag{
		Name:  serverId,
		Usage: "[Optional] Artifactory server ID configured using the config command.` `",
	},
	sshKeyPath: cli.StringFlag{
		Name:  sshKeyPath,
		Usage: "[Optional] SSH key file path.` `",
	},
	sshPassPhrase: cli.StringFlag{
		Name:  sshPassPhrase,
		Usage: "[Optional] SSH key passphrase.` `",
	},
	clientCertPath: cli.StringFlag{
		Name:  clientCertPath,
		Usage: "[Optional] Client certificate file in PEM format.` `",
	},
	clientCertKeyPath: cli.StringFlag{
		Name:  clientCertKeyPath,
		Usage: "[Optional] Private key file for the client certificate in PEM format.` `",
	},
	sortBy: cli.StringFlag{
		Name:  sortBy,
		Usage: "[Optional] A list of semicolon-separated fields to sort by. The fields must be part of the 'items' AQL domain. For more information, see https://www.jfrog.com/confluence/display/RTF/Artifactory+Query+Language#ArtifactoryQueryLanguage-EntitiesandFields` `",
	},
	sortOrder: cli.StringFlag{
		Name:  sortOrder,
		Usage: "[Default: asc] The order by which fields in the 'sort-by' option should be sorted. Accepts 'asc' or 'desc'.` `",
	},
	limit: cli.StringFlag{
		Name:  limit,
		Usage: "[Optional] The maximum number of items to fetch. Usually used with the 'sort-by' option.` `",
	},
	offset: cli.StringFlag{
		Name:  offset,
		Usage: "[Optional] The offset from which to fetch items (i.e. how many items should be skipped). Usually used with the 'sort-by' option.` `",
	},
	spec: cli.StringFlag{
		Name:  spec,
		Usage: "[Optional] Path to a File Spec.` `",
	},
	specVars: cli.StringFlag{
		Name:  specVars,
		Usage: "[Optional] List of variables in the form of \"key1=value1;key2=value2;...\" to be replaced in the File Spec. In the File Spec, the variables should be used as follows: ${key1}.` `",
	},
	buildName: cli.StringFlag{
		Name:  buildName,
		Usage: "[Optional] Providing this option will collect and record build info for this build name. Build number option is mandatory when this option is provided.` `",
	},
	buildNumber: cli.StringFlag{
		Name:  buildNumber,
		Usage: "[Optional] Providing this option will collect and record build info for this build number. Build name option is mandatory when this option is provided.` `",
	},
	module: cli.StringFlag{
		Name:  module,
		Usage: "[Optional] Optional module name for the build-info. Build name and number options are mandatory when this option is provided.` `",
	},
	// Custom usage must be assign to exclude-patterns flag
	excludePatterns: cli.StringFlag{
		Name:   excludePatterns,
		Hidden: true,
	},
	// Custom usage must be assign to exclusions flag
	exclusions: cli.StringFlag{
		Name: exclusions,
	},
	// Custom usage must be assign to exclusions flag
	recursive: cli.BoolTFlag{
		Name: recursive,
	},
	build: cli.StringFlag{
		Name:  build,
		Usage: "[Optional] If specified, only artifacts of the specified build are matched. The property format is build-name/build-number. If you do not specify the build number, the artifacts are filtered by the latest build number.` `",
	},
	includeDirs: cli.BoolFlag{
		Name:  includeDirs,
		Usage: "[Default: false] Set to true if you'd like to also apply the source path pattern for directories and not just for files.` `",
	},
	// Custom usage must be assign to props flag
	props: cli.StringFlag{
		Name:  props,
		Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\".` `",
	},
	// Custom usage must be assign to exclude-props flag
	excludeProps: cli.StringFlag{
		Name:  excludeProps,
		Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\".` `",
	},
	failNoOp: cli.BoolFlag{
		Name:  failNoOp,
		Usage: "[Default: false] Set to true if you'd like the command to return exit code 2 in case of no files are affected.` `",
	},
	// Custom usage must be assign to threads flag
	threads: cli.StringFlag{
		Name:  threads,
		Value: "",
	},
	// Custom usage must be assign to sync-deletes flag
	syncDeletes: cli.StringFlag{
		Name: syncDeletes,
	},
	// Custom usage must be assign to quite flag
	quite: cli.StringFlag{
		Name: quite,
	},
	insecureTls: cli.BoolFlag{
		Name:  insecureTls,
		Usage: "[Default: false] Set to true to skip TLS certificates verification.` `",
	},
	bundle: cli.StringFlag{
		Name:  bundle,
		Usage: "[Optional] If specified, only artifacts of the specified bundle are matched. The value format is bundle-name/bundle-version.` `",
	},
	archiveEntries: cli.StringFlag{
		Name:  archiveEntries,
		Usage: "[Optional] If specified, only archive artifacts containing entries matching this pattern are matched. You can use wildcards to specify multiple artifacts.` `",
	},
	interactive: cli.BoolTFlag{
		Name:  "interactive",
		Usage: "[Default: true, unless $CI is true] Set to false if you do not want the config command to be interactive. If true, the --url option becomes optional.` `",
	},
	encPassword: cli.BoolTFlag{
		Name:  "enc-password",
		Usage: "[Default: true] If set to false then the configured password will not be encrypted using Artifactory's encryption API.` `",
	},
}

var commandFlags = map[string][]string{
	"config": []string{interactive, encPassword, url, distUrl, password, apikey, accessToken, sshKeyPath, clientCertPath, clientCertKeyPath},
}

func GetCommandFlags(cmd string) (flags []cli.Flag) {
	flagList, ok := commandFlags[cmd]
	if !ok {
		return
	}
	for _, flag := range flagList {
		flags = append(flags, flagsMap[flag])
	}
	sort.Slice(flags, func(i, j int) bool { return flags[i].GetName() < flags[j].GetName() })
	return
}
