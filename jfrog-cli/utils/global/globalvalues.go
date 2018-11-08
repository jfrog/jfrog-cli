package global

var global *Global

type Global struct {
	modulesPublished map[string]bool
	success          int
	failures         int
	total            int
}

func (g *Global) GetGlobalMap() map[string]bool {
	g.initMap()
	return g.modulesPublished
}

func (g *Global) GetSuccess() int {
	return g.success
}

func (g *Global) GetFailures() int {
	return g.failures
}

func (g *Global) GetTotal() int {
	return g.total
}

func (g *Global) IncreaseSuccess() {
	g.success += 1
}

func (g *Global) IncreaseFailures() {
	g.failures += 1
}

func (g *Global) IncreaseTotal(sum int) {
	g.total += sum
}

func (g *Global) initMap() {
	if g.modulesPublished == nil {
		g.modulesPublished = make(map[string]bool)
	}
}

func GetGlobalVariables() *Global {
	if global == nil {
		global = &Global{}
	}
	global.initMap()
	return global
}
