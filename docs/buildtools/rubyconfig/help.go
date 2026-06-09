package rubyconfig

var Usage = []string{"ruby-config [command options]"}

func GetDescription() string {
	return "Generate ruby build configuration."
}

func GetAIDescription() string {
	return `Write a per-project Ruby configuration (.jfrog/projects/ruby.yaml) and update bundler's source list so gem resolution routes through an Artifactory RubyGems repository.

When to use:
- Initial setup of a Ruby/Bundler project to use a private RubyGems index.

Prerequisites:
- A configured server.
- The Artifactory RubyGems repository key.

Common patterns:
  $ jf ruby-config --server-id-resolve=my-server --repo-resolve=rubygems-virtual

Gotchas:
- Interactive prompts trigger when required flags are missing.
- 'jf ruby-config' modifies the Gemfile or bundler config; review the diff after running.

Related: jf rt build-publish`
}
