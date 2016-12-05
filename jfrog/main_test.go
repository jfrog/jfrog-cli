package main

import (
	"testing"
	"flag"
	"os"
)

func TestMain(m *testing.M) {
	flag.Parse()
	InitBintrayTests()
	InitArtifactoryTests()

	result := m.Run()

	CleanBintrayTests()
	CleanArtifactoryTests()
	os.Exit(result)
}