package dbo

import "strings"

// NullStrings nullable strings
type NullStrings struct {
	Strings []string
	Valid   bool
}

// ToInterfaceSlice values to interface() slice
func (s NullStrings) ToInterfaceSlice() []interface{} {
	slice := make([]interface{}, len(s.Strings))
	for index, value := range s.Strings {
		slice[index] = value
	}

	return slice
}

// SQLPlaceHolder generate sql place holder
func (s NullStrings) SQLPlaceHolder() string {
	if len(s.Strings) == 0 && s.Valid {
		return "null"
	}

	return strings.TrimSuffix(strings.Repeat("?,", len(s.Strings)), ",")
}

// NullInts nullable ints
type NullInts struct {
	Ints  []int
	Valid bool
}

// ToInterfaceSlice values to interface() slice
func (s NullInts) ToInterfaceSlice() []interface{} {
	slice := make([]interface{}, len(s.Ints))
	for index, value := range s.Ints {
		slice[index] = value
	}

	return slice
}

// SQLPlaceHolder generate sql place holder
func (s NullInts) SQLPlaceHolder() string {
	if len(s.Ints) == 0 && s.Valid {
		return "null"
	}

	return strings.TrimSuffix(strings.Repeat("?,", len(s.Ints)), ",")
}
