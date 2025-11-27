package shared

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"
)

type DataType interface {
	fmt.Stringer
	RootType() string
	ZeroValue() string
}

// Ensure all data types implement fmt.Stringer
func (n Number) String() string  { return Keyword_Number }
func (s String) String() string  { return Keyword_String }
func (b Boolean) String() string { return Keyword_Boolean }
func (l List) String() string    { return fmt.Sprintf("list[%s]", l.ListType.String()) }
func (m Map) String() string {
	return fmt.Sprintf("map[%s][%s]", m.KeyType.RootType(), m.ValueType.String())
}
func (o Object) String() string {
	var sb strings.Builder
	sb.WriteRune('{')
	keys := slices.Collect(maps.Keys(o.Keys))
	sort.Strings(keys)
	for _, key := range keys {
		sb.WriteString(key)
		sb.WriteRune(' ')
		sb.WriteString(o.Keys[key].String())
		sb.WriteRune(',')
	}
	sb.WriteRune('}')
	return sb.String()
}
func (c Custom) String() string { return c.Name }

// Ensure all data types have the RootType function
func (n Number) RootType() string  { return Keyword_Number }
func (s String) RootType() string  { return Keyword_String }
func (b Boolean) RootType() string { return Keyword_Boolean }
func (l List) RootType() string    { return fmt.Sprintf("list[%s]", l.ListType.RootType()) }
func (m Map) RootType() string {
	return fmt.Sprintf("map[%s][%s]", m.KeyType.RootType(), m.ValueType.RootType())
}
func (o Object) RootType() string {
	var sb strings.Builder
	sb.WriteRune('{')
	keys := slices.Collect(maps.Keys(o.Keys))
	sort.Strings(keys)
	for _, key := range keys {
		sb.WriteString(key)
		sb.WriteRune(' ')
		sb.WriteString(o.Keys[key].RootType())
		sb.WriteRune(',')
	}
	sb.WriteRune('}')
	return sb.String()
}
func (c Custom) RootType() string { return c.DataType.RootType() }

// Ensure all data types have the ZeroValue Function
func (n Number) ZeroValue() string  { return "0" }
func (s String) ZeroValue() string  { return `""` }
func (b Boolean) ZeroValue() string { return "false" }
func (l List) ZeroValue() string    { return "[]" }
func (m Map) ZeroValue() string     { return "{}" }
func (o Object) ZeroValue() string  { return "{}" }
func (c Custom) ZeroValue() string  { return c.DataType.ZeroValue() }

type Number struct{}
type String struct{}
type Boolean struct{}

type List struct {
	ListType DataType
}

type Map struct {
	KeyType   DataType
	ValueType DataType
}

type Object struct {
	Keys map[string]DataType
}

type Custom struct {
	Name     string
	DataType DataType
}

func DataTypeFromString(input string) DataType {
	switch input {
	case Keyword_Number:
		return Number{}
	case Keyword_String:
		return String{}
	case Keyword_Boolean:
		return Boolean{}
	default:
		return Custom{Name: input}
	}
}
