package dbo

import "time"

type NullStrings struct {
	Strings []string
	Valid   bool
}

type NullString struct {
	String string
	Valid  bool
}

type NullInt struct {
	Int   int
	Valid bool
}

type NullInts struct {
	Ints  []int
	Valid bool
}

type NullInt64 struct {
	Int64 int64
	Valid bool
}

type NullInt64s struct {
	Int64s []int64
	Valid  bool
}

type NullBool struct {
	Bool  bool
	Valid bool
}

type NullBools struct {
	Bools []bool
	Valid bool
}

type NullTime struct {
	Time  time.Time
	Valid bool
}

type NullTimes struct {
	Times []time.Time
	Valid bool
}

func Var[T any](v T) *T {
	return &v
}
