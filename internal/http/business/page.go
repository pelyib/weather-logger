package business

type Page struct {
	Title       string
	Breadcrumbs [][]Breadcrumb
	Chart       Chart
}

type Breadcrumb struct {
	Title      string
	UriPart    string
	UriValue   string
	IsSelected bool
}

func MakeBreadcrumb(t string, up string, uv string, isSelected bool) Breadcrumb {
	return Breadcrumb{Title: t, UriPart: up, UriValue: uv, IsSelected: isSelected}
}
