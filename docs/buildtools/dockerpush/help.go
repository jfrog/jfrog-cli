package dockerpush

var Usage = []string{"docker push <image tag> [command options]"}

func GetDescription() string {
	return `Run Docker push command.`
}

func GetArguments() string {
	return `	docker push args
		The docker push args to run docker push.
		
	--validate-sha
		Set to true to enable SHA-based validation during Docker push.
		When enabled, manifest validation will use the image's SHA digest instead of name:tag.
		This is useful when pushing to virtual repositories where the tag might exist with different content in higher priority repositories.
		The SHA digest is automatically determined from the local image.
		
		With this flag, the CLI will:
		1. Use the local image's SHA digest for validation instead of the tag
		2. Attempt to find the image in the repository by SHA if tag-based lookup fails
		3. Continue the operation even if a digest mismatch is detected, with appropriate warnings`
}
