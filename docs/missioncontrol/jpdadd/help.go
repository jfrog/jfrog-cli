package jpdadd

var Usage = []string{"mc ja [command options] <config>"}

func GetDescription() string {
	return "Add a JPD to Mission Control."
}

func GetArguments() string {
	return `	Config
		Path to a JSON configuration file containing the JPD details.`
}

func GetAIDescription() string {
	return `Register a JFrog Platform Deployment (JPD) with a Mission Control server. The JPD definition is read from a local JSON file (URL, location, type, token, etc.).

When to use:
- Onboarding a new platform deployment to a Mission Control fleet.
- Scripting JPD registration during platform provisioning.

Prerequisites:
- A configured Mission Control server (jf c add captures the mission control URL).
- Admin privileges on the Mission Control side.
- A JSON file matching the JPD schema (name, url, location, token, type).

Common patterns:
  $ jf mc ja ./jpd-config.json
  $ jf mc ja ./jpd-config.json --format=json

Gotchas:
- The JSON file must match the JPD schema exactly; common omissions are 'location' and 'token'.
- The Mission Control URL must be set in the active server configuration; missing URL produces a confusing error.

Related: jf mc jd, jf mc la, jf mc ld, jf mc lr`
}
