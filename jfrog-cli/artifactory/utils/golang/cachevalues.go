package golang

type DependenciesCache struct {
	modulesPublished map[string]bool
	successes        int
	failures         int
	total            int
}

func (dc *DependenciesCache) GetMap() map[string]bool {
	dc.initMap()
	return dc.modulesPublished
}

func (dc *DependenciesCache) GetSuccesses() int {
	return dc.successes
}

func (dc *DependenciesCache) GetFailures() int {
	return dc.failures
}

func (dc *DependenciesCache) GetTotal() int {
	return dc.total
}

func (dc *DependenciesCache) IncrementSuccess() {
	dc.successes += 1
}

func (dc *DependenciesCache) IncrementFailures() {
	dc.failures += 1
}

func (dc *DependenciesCache) IncrementTotal(sum int) {
	dc.total += sum
}

func (dc *DependenciesCache) initMap() {
	if dc.modulesPublished == nil {
		dc.modulesPublished = make(map[string]bool)
	}
}