package vcs

import "strings"

type technology string

const (
	maven  = "maven"
	gradle = "gradle"
	npm    = "npm"
)

type TechnologyIndicator interface {
	GetTechnology() technology
	Indicates(file string) bool
}

type MavenIndicator struct {
}

func (mi MavenIndicator) GetTechnology() technology {
	return technology(maven)
}

func (mi MavenIndicator) Indicates(file string) bool {
	return strings.Contains(file, "pom.xml")
}

type GradleIndicator struct {
}

func (gi GradleIndicator) GetTechnology() technology {
	return technology(gradle)
}

func (gi GradleIndicator) Indicates(file string) bool {
	return strings.Contains(file, ".gradle")
}

type NpmIndicator struct {
}

func (ni NpmIndicator) GetTechnology() technology {
	return technology(npm)
}

func (ni NpmIndicator) Indicates(file string) bool {
	return strings.Contains(file, "package.json")
}

func GetTechIndicators() []TechnologyIndicator {
	return []TechnologyIndicator{
		MavenIndicator{},
		GradleIndicator{},
		NpmIndicator{},
	}
}
