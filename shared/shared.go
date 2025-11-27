package shared

const (
	Keyword_Function = "fn"
	Keyword_Number   = "num"
	Keyword_String   = "str"
	Keyword_Boolean  = "bool"
	Keyword_List     = "list"
	Keyword_Map      = "map"
	Keyword_Object   = "obj"
	Keyword_True     = "true"
	Keyword_False    = "false"
	Keyword_Local    = "local"
)

var (
	IllegalKeywords = map[string]struct{}{
		Keyword_Function: {},
		Keyword_Number:   {},
		Keyword_String:   {},
		Keyword_Boolean:  {},
		Keyword_List:     {},
		Keyword_Map:      {},
		Keyword_True:     {},
		Keyword_False:    {},
	}
	BuiltInDataTypes = map[string]struct{}{
		Keyword_Number:  {},
		Keyword_String:  {},
		Keyword_Boolean: {},
	}
	DataTypesZeroValues = map[string]string{
		Keyword_Number:  "0",
		Keyword_String:  `""`,
		Keyword_Boolean: "false",
	}
)
