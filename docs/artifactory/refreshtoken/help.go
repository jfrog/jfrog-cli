package refreshtoken

const Description = "Refresh access token."

var Usage = []string{"jfrog rt token-refresh [command options] <refresh token> <access token>"}

const Arguments string = `	refresh token
		The refresh token used for refreshing the access token.

	access token
		The access token to refresh.`
