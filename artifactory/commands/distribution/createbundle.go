package distribution

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/utils/config"
)

type ReleaseNotesType string

const (
	Markdown  ReleaseNotesType = "markdown"
	Asciidoc                   = "asciidoc"
	PlainText                  = "plain_text"
)

type CreateBundleCommand struct {
	name             string
	version          string
	dryRun           bool
	signImmediately  bool
	description      string
	releaseNotesPath string
	releaseNotesType ReleaseNotesType
	spec             *spec.SpecFiles
	rtDetails        *config.ArtifactoryDetails
}

func NewCreateBundleCommand() *CreateBundleCommand {
	return &CreateBundleCommand{}
}

func (cbc *CreateBundleCommand) SetName(name string) *CreateBundleCommand {
	cbc.name = name
	return cbc
}

func (cbc *CreateBundleCommand) SetVersion(version string) *CreateBundleCommand {
	cbc.version = version
	return cbc
}

func (cbc *CreateBundleCommand) SetDryRun(dryRun bool) *CreateBundleCommand {
	cbc.dryRun = dryRun
	return cbc
}

func (cbc *CreateBundleCommand) SetSignImmediately(signImmediately bool) *CreateBundleCommand {
	cbc.signImmediately = signImmediately
	return cbc
}

func (cbc *CreateBundleCommand) SetDescription(description string) *CreateBundleCommand {
	cbc.description = description
	return cbc
}

func (cbc *CreateBundleCommand) SetReleaseNotesPath(releaseNotesPath string) *CreateBundleCommand {
	cbc.releaseNotesPath = releaseNotesPath
	return cbc
}

func (cbc *CreateBundleCommand) SetReleaseNotesType(releaseNotesType ReleaseNotesType) *CreateBundleCommand {
	cbc.releaseNotesType = releaseNotesType
	return cbc
}

func (cbc *CreateBundleCommand) SetSpec(spec *spec.SpecFiles) *CreateBundleCommand {
	cbc.spec = spec
	return cbc
}