package permissiontargetdelete

var Usage = []string{"rt ptdel <permission target name>"}

func GetDescription() string {
	return "Permanently delete a permission target."
}

func GetArguments() string {
	return `	permission target name
		Specifies the permission target that should be removed.`
}
