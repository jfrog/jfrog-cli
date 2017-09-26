package utils

const ClassworldsConf =
`main is org.apache.maven.cli.MavenCli from plexus.core

set maven.home default ${user.home}/m2

[plexus.core]
load ${maven.home}/lib/*.jar
load ${m3plugin.lib}/*.jar
`

