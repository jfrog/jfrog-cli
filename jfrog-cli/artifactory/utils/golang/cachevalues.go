package golang

var cache DynamicCache

type DynamicCache struct {
	modulesPublished map[string]bool
	success          int
	failures         int
	total            int
}

func (dc *DynamicCache) GetGlobalMap() map[string]bool {
	dc.initMap()
	return dc.modulesPublished
}

func (dc *DynamicCache) GetSuccess() int {
	return dc.success
}

func (dc *DynamicCache) GetFailures() int {
	return dc.failures
}

func (dc *DynamicCache) GetTotal() int {
	return dc.total
}

func (dc *DynamicCache) IncreaseSuccess() {
	dc.success += 1
}

func (dc *DynamicCache) IncreaseFailures() {
	dc.failures += 1
}

func (dc *DynamicCache) IncreaseTotal(sum int) {
	dc.total += sum
}

func (dc *DynamicCache) initMap() {
	if dc.modulesPublished == nil {
		dc.modulesPublished = make(map[string]bool)
	}
}

func GetStaticCache() DynamicCache {
	if &cache == nil {
		cache = DynamicCache{}
	}
	cache.initMap()
	return cache
}
