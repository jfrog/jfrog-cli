package transfersettings

var Usage = []string{"rt transfer-settings"}

func GetDescription() string {
	return "Configure the settings for the 'jf rt transfer-files' command."
}

func GetAIDescription() string {
	return `Interactively tune throughput settings (max parallel uploads, etc.) for 'jf rt transfer-files'. Settings are saved under ~/.jfrog/ and reused by subsequent transfer-files runs.

When to use:
- Before kicking off a large transfer, to set realistic parallelism for the source server's capacity.
- Tuning ongoing transfers if throughput is too low or saturating the source.

Prerequisites:
- A configured Artifactory server.

Common patterns:
  $ jf rt transfer-settings

Gotchas:
- Interactive only; cannot be scripted directly. For automation, edit ~/.jfrog/transfer-settings.json directly.
- Settings apply across all subsequent transfer-files invocations.

Related: jf rt transfer-files, jf rt transfer-config`
}
