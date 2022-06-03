package business

type Page struct {
	Title       string
	Breadcrumbs [][]Breadcrumb
	Chart       Chart
}

type Breadcrumb struct {
	Title      string
	Link       string
	IsSelected bool
}
