package workspacerun

var Usage = []string{"pl wsr"}

func GetDescription() string {
	return "Trigger workspace validation, sync and pipelines run, depending on success on previous stages."
}
