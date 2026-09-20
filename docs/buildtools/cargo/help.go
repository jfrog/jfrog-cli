package cargo

var Usage = []string{"cargo <cargo args> [command options]"}

func GetDescription() string {
	return "Run Cargo (Rust) commands with JFrog build-info collection."
}

func GetArguments() string {
	return `	cargo args
		Cargo subcommand and its native flags — passed through to the local cargo binary
		verbatim (e.g. build, publish, install, package, add). Only the JFrog-specific flags
		below (--build-name, --build-number, --module, --project, --server-id) are consumed by
		jf; everything else after 'cargo' goes to cargo unchanged.`
}

func GetAIDescription() string {
	return `Run a Cargo (Rust) build with JFrog instrumentation: resolves crates through an Artifactory cargo remote, optionally publishes to a cargo local repo, and collects a build-info record for traceability. Wraps the local cargo binary; the subcommand and cargo-native flags are passed through, while a small set of jf-only flags govern build-info collection.

When to use:
- Building or publishing Rust projects against Artifactory cargo repositories.
- Producing a JFrog build-info record alongside a cargo build / install / publish.

Prerequisites:
- A local cargo binary on PATH (this command does not install cargo).
- 'jf setup cargo' run once (writes ~/.cargo/config.toml with the jfrog + jfrog-local registries).
- A configured server (jf c add) that setup can reference.

Common patterns:
  $ jf cargo build --build-name=my-build --build-number=1
  $ jf cargo publish -p my-crate --registry jfrog-local --build-name=my-build --build-number=1
  $ jf cargo install --path . --features tls

Gotchas:
- 'jf setup cargo' must be run first; without it the resolve/publish registries are unknown.
- --build-name and --build-number are required together for build-info collection; passing only one is silently ignored.
- All flags after 'cargo' are passed verbatim to cargo; jf-specific flags (--build-name, --build-number, --module, --project, --server-id) are consumed by jf, the rest go to cargo.

Related: jf setup cargo, jf rt build-publish`
}
