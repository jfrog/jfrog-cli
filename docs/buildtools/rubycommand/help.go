package rubycommand

var Usage = []string{"ruby <gem|bundle> <args> [command options]"}

func GetDescription() string {
	return "Run native RubyGems (gem) and Bundler (bundle) commands with Artifactory authentication and build-info support."
}

func GetAIDescription() string {
	return `Run a native gem or bundle command through JFrog CLI so dependencies resolve against Artifactory and build-info can be recorded. Arguments are forwarded to the native tool unchanged; only --build-name, --build-number, --module, --project, --server-id, and --repo are interpreted by jf.

When to use:
- Installing gems from an Artifactory-backed RubyGems repository (gem install, bundle install).
- Pushing gems to Artifactory (gem push).
- Capturing build-info for a Ruby build by passing --build-name and --build-number.

Prerequisites:
- Ruby, gem, and/or bundle installed and available on PATH.
- A configured JFrog Platform server (jf c add or jf login), or pass --server-id.

Common patterns:
  $ jf ruby gem install rake --repo gems-virtual
  $ jf ruby bundle install --build-name=my-app --build-number=1
  $ jf ruby gem push my_gem-1.0.0.gem --repo gems-local --build-name=my-app --build-number=1
  $ jf ruby bundle install --server-id my-rt --repo gems-virtual

Gotchas:
- All arguments after the tool name (gem/bundle) are passed straight to the native tool (SkipFlagParsing).
- Build-info is collected only when both --build-name and --build-number are provided; publish it afterwards with 'jf rt build-publish'.
- The --repo flag constructs the full Artifactory gems API URL from your server config, so you don't need to pass full URLs.

Related: jf rt build-publish, jf ruby-config`
}

func GetArguments() string {
	return `	ruby <gem|bundle> <args>
		Wraps the native 'gem' and 'bundle' tools. The first argument selects the
		native tool; everything after it is passed straight through. Only
		--build-name, --build-number, --module, --project, --server-id and --repo
		are interpreted by jf.

		Authentication is injected automatically from your jf server config and
		respects credentials you have already configured natively (Gemfile source,
		.bundle/config, ~/.gem/credentials, BUNDLE_* / GEM_HOST_API_KEY env vars).

		The --repo flag specifies the Artifactory repository name and constructs
		the full URL from your server config (eliminating the need to pass full
		Artifactory URLs). When --repo is used with gem install/push, the
		--source/--host argument is injected automatically.

		Examples:
		- jf ruby bundle install --build-name=my-build --build-number=1
		- jf ruby bundle update rake --server-id=my-rt
		- jf ruby gem install rails --repo gems-virtual
		- jf ruby gem install rails --source https://server/artifactory/api/gems/gems-remote/
		- jf ruby gem push my_gem-1.0.0.gem --repo gems-local --build-name=my-build --build-number=1
		- jf ruby gem push my_gem-1.0.0.gem --host https://server/artifactory/api/gems/gems-local/ --build-name=my-build --build-number=1
		- jf ruby gem install rake --repo gems-virtual --server-id my-rt`
}
