package pipinstall

const Description = "Run pip install."

var Usage = []string{`jfrog rt pipi [command options] <install file>`}

const Arguments string = `	install file
		The project installation file for pip. Usually 'setup.py' or 'requirements.txt'.`
