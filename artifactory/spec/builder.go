package spec

import "strconv"

type builder struct {
	pattern         string
	// Deprecated, use Exclusions instead
	excludePatterns []string
	exclusions      []string
	target          string
	explode         string
	props           string
	excludeProps	string
	sortOrder       string
	sortBy          []string
	offset          int
	limit           int
	build           string
	recursive       bool
	flat            bool
	regexp          bool
	includeDirs     bool
	archiveEntries  string
}

func NewBuilder() *builder {
	return &builder{}
}

func (b *builder) Pattern(pattern string) *builder {
	b.pattern = pattern
	return b
}

func (b *builder) ArchiveEntries(archiveEntries string) *builder {
	b.archiveEntries = archiveEntries
	return b
}

func (b *builder) ExcludePatterns(excludePatterns []string) *builder {
	b.excludePatterns = excludePatterns
	return b
}

func (b *builder) Exclusions(exclusions []string) *builder {
	b.exclusions = exclusions
	return b
}

func (b *builder) Target(target string) *builder {
	b.target = target
	return b
}

func (b *builder) Explode(explode string) *builder {
	b.explode = explode
	return b
}

func (b *builder) Props(props string) *builder {
	b.props = props
	return b
}

func (b *builder) ExcludeProps(excludeProps string) *builder {
	b.excludeProps = excludeProps
	return b
}

func (b *builder) SortOrder(sortOrder string) *builder {
	b.sortOrder = sortOrder
	return b
}

func (b *builder) SortBy(sortBy []string) *builder {
	b.sortBy = sortBy
	return b
}

func (b *builder) Offset(offset int) *builder {
	b.offset = offset
	return b
}

func (b *builder) Limit(limit int) *builder {
	b.limit = limit
	return b
}

func (b *builder) Build(build string) *builder {
	b.build = build
	return b
}

func (b *builder) Recursive(recursive bool) *builder {
	b.recursive = recursive
	return b
}

func (b *builder) Flat(flat bool) *builder {
	b.flat = flat
	return b
}

func (b *builder) Regexp(regexp bool) *builder {
	b.regexp = regexp
	return b
}

func (b *builder) IncludeDirs(includeDirs bool) *builder {
	b.includeDirs = includeDirs
	return b
}

func (b *builder) BuildSpec() *SpecFiles {
	return &SpecFiles{
		Files: []File{
			{
				Pattern:         b.pattern,
				// Deprecated, use Exclusions instead
				ExcludePatterns: b.excludePatterns,
				Exclusions:      b.exclusions,
				Target:          b.target,
				Props:           b.props,
				ExcludeProps:	 b.excludeProps,
				SortOrder:       b.sortOrder,
				SortBy:          b.sortBy,
				Offset:          b.offset,
				Limit:           b.limit,
				Build:           b.build,
				Explode:         b.explode,
				Recursive:       strconv.FormatBool(b.recursive),
				Flat:            strconv.FormatBool(b.flat),
				Regexp:          strconv.FormatBool(b.regexp),
				IncludeDirs:     strconv.FormatBool(b.includeDirs),
				ArchiveEntries:  b.archiveEntries,
			},
		},
	}
}
