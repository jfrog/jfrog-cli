package commands

import "strings"

type Technology string

const (
	Maven  = "maven"
	Gradle = "gradle"
	Npm    = "npm"
)

type TechnologyIndicator interface {
	GetTechnology() Technology
	Indicates(file string) bool
}

type MavenIndicator struct {
}

func (mi MavenIndicator) GetTechnology() Technology {
	return Technology(Maven)
}

func (mi MavenIndicator) Indicates(file string) bool {
	return strings.Contains(file, "pom.xml")
}

type GradleIndicator struct {
}

func (gi GradleIndicator) GetTechnology() Technology {
	return Technology(Gradle)
}

func (gi GradleIndicator) Indicates(file string) bool {
	return strings.Contains(file, ".gradle")
}

type NpmIndicator struct {
}

func (ni NpmIndicator) GetTechnology() Technology {
	return Technology(Npm)
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
