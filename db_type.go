package dbo

type DBType string

const (
	MySQL DBType = "mysql"
)

func (t DBType) String() string {
	return string(t)
}

func (t DBType) DriverName() string {
	switch t {
	case MySQL:
		return "mysql"
	default:
		return ""
	}
}
