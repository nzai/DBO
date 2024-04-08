package dbo

type QueryCondition interface {
	GetConditions() ([]string, []any)
}

type OrderByCondition interface {
	GetOrderBy() string
}

type PagerCondition interface {
	GetPager() *Pager
}
