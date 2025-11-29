package compiler

import (
	"errors"
	"fmt"
	"pixie/parser"
	"pixie/shared"
	"strings"
)

const (
	globalScope = 1
)

var (
	ErrInvalidTypeAssign = fmt.Errorf("invalid type assign")
)

func Compile(node parser.Node) (lua string, err error) {
	stmt, ok := node.(parser.Stmt)
	if !ok {
		err = fmt.Errorf("expected statement, got: %v", node)
		return
	}

	var sb strings.Builder
	c := &compiler{
		sb:        &sb,
		variables: make(map[string]variable, 0),
		objects:   make(map[string]object, 0),
	}
	if err = c.compileStmt(stmt); err != nil {
		err = fmt.Errorf("failed to compile statement: %w", err)
		return
	}
	return sb.String(), nil
}

type compiler struct {
	sb        *strings.Builder
	scope     int
	variables map[string]variable
	objects   map[string]object
}

type variable struct {
	scope    int
	dataType shared.DataType
}

type object struct {
	fields []parser.FieldTypePair
}

func (c *compiler) compileStmt(stmt parser.Stmt) (err error) {
	switch n := stmt.(type) {
	case parser.StmtBlock:
		if err = c.compileStmtBlock(n); err != nil {
			err = fmt.Errorf("failed to compile statement block: %w", err)
			return
		}
	case parser.StmtCallFunction:
		if err = c.compileStmtCallFunction(n); err != nil {
			err = fmt.Errorf("failed to compile statement call function: %w", err)
			return
		}
	case parser.StmtVarDeclare:
		if err = c.compileStmtVarDeclare(n); err != nil {
			err = fmt.Errorf("failed to compile statement variable declare: %w", err)
			return
		}
	case parser.StmtVarAssign:
		if err = c.compileStmtVarAssign(n); err != nil {
			err = fmt.Errorf("failed to compile statement assign: %w", err)
			return
		}
	case parser.StmtObjDefine:
		if err = c.compileStmtObjDefine(n); err != nil {
			err = fmt.Errorf("failed to compile statement object define: %w", err)
			return
		}
	default:
		err = fmt.Errorf("expected statement, got: %v", n)
		return
	}

	return nil
}

func (c *compiler) compileExpr(expr parser.Expr) (err error) {
	switch n := expr.(type) {
	case parser.ExprBlock:
		if err = c.compileExprBlock(n); err != nil {
			err = fmt.Errorf("failed to compile expression block: %w", err)
			return
		}
	case parser.ExprNumber:
		if err = c.compileExprNumber(n); err != nil {
			err = fmt.Errorf("failed to compile expression number: %w", err)
			return
		}
	case parser.ExprString:
		if err = c.compileExprString(n); err != nil {
			err = fmt.Errorf("failed to compile expression string: %w", err)
			return
		}
	case parser.ExprBoolean:
		if err = c.compileExprBoolean(n); err != nil {
			err = fmt.Errorf("failed to compile expression boolean: %w", err)
			return
		}
	case parser.ExprList:
		if err = c.compileExprList(n); err != nil {
			err = fmt.Errorf("failed to compile expression list: %w", err)
			return
		}
	case parser.ExprTable:
		if err = c.compileExprTable(n); err != nil {
			err = fmt.Errorf("failed to compile expression table: %w", err)
			return
		}
	case parser.ExprVariable:
		if err = c.compileExprVariable(n); err != nil {
			err = fmt.Errorf("failed to compile expression variable: %w", err)
			return
		}
	case parser.ExprIndex:
		if err = c.compileExprIndex(n); err != nil {
			err = fmt.Errorf("failed to compile expression index: %w", err)
			return
		}
	case parser.ExprPropertyAccess:
		if err = c.compileExprPropertyAccess(n); err != nil {
			err = fmt.Errorf("failed to compile expression property access: %w", err)
			return
		}
	default:
		err = fmt.Errorf("expected expr, got: %v", n)
		return
	}

	return nil
}

func (c *compiler) compileStmtBlock(stmt parser.StmtBlock) (err error) {
	c.scope += 1
	for _, s := range stmt.Stmts {
		err = c.compileStmt(s)
		if err != nil {
			err = fmt.Errorf("failed to compile stmt: %w", err)
			return
		}
		c.sb.WriteRune('\n')
	}

	variablesToRemove := make([]string, 0, len(c.variables))
	for k, v := range c.variables {
		if v.scope == c.scope {
			variablesToRemove = append(variablesToRemove, k)
		}
	}

	for _, name := range variablesToRemove {
		delete(c.variables, name)
	}

	c.scope -= 1
	return nil
}

func (c *compiler) compileCommaSeparatedExpressions(exprs []parser.Expr) (err error) {
	argsLen := len(exprs)
	for i, arg := range exprs {
		if err = c.compileExpr(arg); err != nil {
			err = fmt.Errorf("failed to compile argument %d: %w", i, err)
			return
		}

		if i < argsLen-1 {
			c.sb.WriteRune(',')
		}
	}
	return nil
}

func (c *compiler) compileStmtCallFunction(stmt parser.StmtCallFunction) (err error) {
	c.sb.WriteString(stmt.FunctionName)
	c.sb.WriteRune('(')
	if err = c.compileCommaSeparatedExpressions(stmt.Args); err != nil {
		err = fmt.Errorf("failed to compile comma separated expressions: %w", err)
		return
	}
	c.sb.WriteRune(')')
	return nil
}

func (c *compiler) compileStmtVarDeclare(stmt parser.StmtVarDeclare) (err error) {
	_, ok := c.variables[stmt.VariableName]
	if ok {
		err = fmt.Errorf("variable %q already exists", stmt.VariableName)
		return
	}

	variable := variable{
		scope:    c.scope,
		dataType: stmt.DataType,
	}
	c.variables[stmt.VariableName] = variable

	// Check if we need to declare a local variable
	if c.scope != globalScope {
		c.sb.WriteString(shared.Keyword_Local)
		c.sb.WriteRune(' ')
	}

	// Write the variable name
	c.sb.WriteString(stmt.VariableName)
	c.sb.WriteString(" = ")

	// Write the expression
	if stmt.Expr == nil {
		if err = c.compileDataTypeZeroValue(variable.dataType); err != nil {
			err = fmt.Errorf("failed to compile data type zero value: %w", err)
			return
		}
	} else {
		// Check if this is an incomplete object assignment that needs to be filled with zero values
		if exprTable, isTable := stmt.Expr.(parser.ExprTable); isTable {
			if customType, isCustom := variable.dataType.(shared.Custom); isCustom {
				if _, isObject := c.objects[customType.Name]; isObject {
					// This is an assignment of a table to a custom object variable
					// We need to handle incomplete object assignments by filling missing fields
					if err = c.compileExprTableWithZeroValues(exprTable, customType); err != nil {
						err = fmt.Errorf("failed to compile expression table with zero values: %w", err)
						return
					}
					return nil
				}
			}
		}

		if err = c.compileExpr(stmt.Expr); err != nil {
			err = fmt.Errorf("failed to parse expression: %w", err)
			return
		}
	}

	return nil
}

func (c *compiler) compileStmtVarAssign(stmt parser.StmtVarAssign) (err error) {
	v, ok := c.variables[stmt.VariableName]
	if !ok {
		err = fmt.Errorf("variable %q does not exist", stmt.VariableName)
		return
	}

	switch e := stmt.Expr.(type) {
	case parser.ExprVariable:
		ev, ok := c.variables[e.Name]
		if !ok {
			err = fmt.Errorf("variable %q does not exist", e.Name)
			return
		}
		if v.dataType.String() != ev.dataType.String() {
			err = errors.Join(ErrInvalidTypeAssign, fmt.Errorf("wanted %q got %q", v.dataType.String(), ev.dataType.String()))
			return
		}
	default:
		if err = c.checkExpressionValidDataType(v.dataType, stmt.Expr); err != nil {
			err = errors.Join(ErrInvalidTypeAssign, fmt.Errorf("%s", err.Error())) // for some reason it wouldn't show the second error when I joined it with the err variable
			return
		}
	}

	c.sb.WriteString(stmt.VariableName)
	c.sb.WriteString(" = ")

	// Check if this is an incomplete object assignment that needs to be filled with zero values
	if exprTable, isTable := stmt.Expr.(parser.ExprTable); isTable {
		if customType, isCustom := v.dataType.(shared.Custom); isCustom {
			if _, isObject := c.objects[customType.Name]; isObject {
				// This is an assignment of a table to a custom object variable
				// We need to handle incomplete object assignments by filling missing fields
				if err = c.compileExprTableWithZeroValues(exprTable, customType); err != nil {
					err = fmt.Errorf("failed to compile expression table with zero values: %w", err)
					return
				}
				return nil
			}
		}
	}

	if err = c.compileExpr(stmt.Expr); err != nil {
		err = fmt.Errorf("failed to parse expression: %w", err)
		return
	}

	return nil
}

func (c *compiler) compileStmtObjDefine(stmt parser.StmtObjDefine) (err error) {
	if _, ok := c.objects[stmt.Name]; ok {
		err = fmt.Errorf("object definition %q already exists", stmt.Name)
		return
	}
	c.objects[stmt.Name] = object{
		fields: stmt.Fields,
	}
	return nil
}

func (c *compiler) compileExprBlock(expr parser.ExprBlock) (err error) {
	c.sb.WriteRune('(')
	if err = c.compileExpr(expr.Value); err != nil {
		err = fmt.Errorf("failed to compile expression: %w", err)
		return
	}
	c.sb.WriteRune(')')
	return nil
}

func (c *compiler) compileExprNumber(expr parser.ExprNumber) (err error) {
	c.sb.WriteString(expr.Value)
	return nil
}

func (c *compiler) compileExprString(expr parser.ExprString) (err error) {
	c.sb.WriteRune('"')
	c.sb.WriteString(expr.Value)
	c.sb.WriteRune('"')
	return nil
}

func (c *compiler) compileExprBoolean(expr parser.ExprBoolean) (err error) {
	c.sb.WriteString(expr.Value)
	return nil
}

func (c *compiler) compileExprList(expr parser.ExprList) (err error) {
	c.sb.WriteRune('[')
	if err = c.compileCommaSeparatedExpressions(expr.Values); err != nil {
		err = fmt.Errorf("failed to compile comma separated expressions: %w", err)
		return
	}
	c.sb.WriteRune(']')
	return nil
}
func (c *compiler) compileExprTable(expr parser.ExprTable) (err error) {
	c.sb.WriteRune('{')
	argsLen := len(expr.Pairs)
	for i, pair := range expr.Pairs {
		// Here I'm making the explicit decision that when creating a table the keys must be primitives (and labels in the case of when this is used to create objects, in which case the label is wrapped in double quotes). I'm doing this to reduce the complexity of the language and compiler, and also because both maps and objects use an underlying table expression and it was going to be hell trying to separate the underlying expressions while maintaining a clean syntax.
		switch k := pair.Key.(type) {
		case parser.ExprNumber:
			if err = c.compileExprNumber(k); err != nil {
				err = fmt.Errorf("failed to compile key: %w", err)
				return
			}
		case parser.ExprString:
			if err = c.compileExprString(k); err != nil {
				err = fmt.Errorf("failed to compile key: %w", err)
				return
			}
		case parser.ExprBoolean:
			if err = c.compileExprBoolean(k); err != nil {
				err = fmt.Errorf("failed to compile key: %w", err)
				return
			}
		case parser.ExprVariable:
			c.sb.WriteRune('"')
			if err = c.compileExprVariable(k); err != nil {
				err = fmt.Errorf("failed to compile key: %w", err)
				return
			}
			c.sb.WriteRune('"')
		}

		c.sb.WriteRune(':')

		if err = c.compileExpr(pair.Value); err != nil {
			err = fmt.Errorf("failed to compile value %d: %w", i, err)
			return
		}

		if i < argsLen-1 {
			c.sb.WriteRune(',')
		}
	}
	c.sb.WriteRune('}')
	return nil
}

func (c *compiler) compileExprTableWithZeroValues(expr parser.ExprTable, objectType shared.Custom) (err error) {
	obj, ok := c.objects[objectType.Name]
	if !ok {
		err = fmt.Errorf("object %q does not exist", objectType.Name)
		return
	}

	// Create a map of provided fields for quick lookup
	providedFields := make(map[string]parser.Expr)
	for _, pair := range expr.Pairs {
		if varExpr, isVar := pair.Key.(parser.ExprVariable); isVar {
			providedFields[varExpr.Name] = pair.Value
		}
	}

	// Write the table with all fields - provided fields with their values, missing fields with zero values
	c.sb.WriteRune('{')
	fieldCount := len(obj.fields)
	for i, field := range obj.fields {
		// Write the field name
		c.sb.WriteRune('"')
		c.sb.WriteString(field.Field)
		c.sb.WriteRune('"')
		c.sb.WriteRune(':')

		// Check if this field was provided in the assignment
		if value, exists := providedFields[field.Field]; exists {
			// Field was provided, use its value
			// If the value is a table and the field type is a custom object,
			// we need to handle it with zero values too for nested incomplete objects
			if tableValue, isTable := value.(parser.ExprTable); isTable {
				if customType, isCustom := field.Type.(shared.Custom); isCustom {
					if _, isObject := c.objects[customType.Name]; isObject {
						// This is an incomplete nested object assignment that needs to be filled with zero values
						if err = c.compileExprTableWithZeroValues(tableValue, customType); err != nil {
							err = fmt.Errorf("failed to compile nested table with zero values for field %q: %w", field.Field, err)
							return
						}
					} else {
						// Not an object, compile normally
						if err = c.compileExpr(value); err != nil {
							err = fmt.Errorf("failed to compile value for field %q: %w", field.Field, err)
							return
						}
					}
				} else {
					// Not a custom type, compile normally
					if err = c.compileExpr(value); err != nil {
						err = fmt.Errorf("failed to compile value for field %q: %w", field.Field, err)
						return
					}
				}
			} else {
				// Not a table expression, compile normally
				if err = c.compileExpr(value); err != nil {
					err = fmt.Errorf("failed to compile value for field %q: %w", field.Field, err)
					return
				}
			}
		} else {
			// Field was not provided, use zero value
			if err = c.compileDataTypeZeroValue(field.Type); err != nil {
				err = fmt.Errorf("failed to compile zero value for field %q: %w", field.Field, err)
				return
			}
		}

		if i < fieldCount-1 {
			c.sb.WriteRune(',')
		}
	}
	c.sb.WriteRune('}')
	return nil
}

func (c *compiler) compileExprVariable(expr parser.ExprVariable) (err error) {
	c.sb.WriteString(expr.Name)
	return nil
}

func (c *compiler) checkExpressionValidDataType(dataType shared.DataType, expr parser.Expr) (err error) {
	switch d := dataType.(type) {
	case shared.Number:
		return c.checkExpressionValidNumber(d, expr)
	case shared.String:
		return c.checkExpressionValidString(d, expr)
	case shared.Boolean:
		return c.checkExpressionValidBoolean(d, expr)
	case shared.List:
		return c.checkExpressionValidList(d, expr)
	case shared.Map:
		return c.checkExpressionValidMap(d, expr)
	case shared.Custom:
		return c.checkExpressionValidCustom(d, expr)
	}

	err = fmt.Errorf("unknown type: %#v", dataType)
	return
}

func (c *compiler) checkExpressionValidNumber(dataType shared.Number, expr parser.Expr) (err error) {
	switch e := expr.(type) {
	case parser.ExprNumber:
		return nil
	default:
		return fmt.Errorf("expected %s got %T", dataType.String(), e)
	}
}

func (c *compiler) checkExpressionValidString(dataType shared.String, expr parser.Expr) (err error) {
	switch e := expr.(type) {
	case parser.ExprString:
		return nil
	default:
		return fmt.Errorf("expected %s got %T", dataType.String(), e)
	}
}

func (c *compiler) checkExpressionValidBoolean(dataType shared.Boolean, expr parser.Expr) (err error) {
	switch e := expr.(type) {
	case parser.ExprBoolean:
		return nil
	default:
		return fmt.Errorf("expected %s got %T", dataType.String(), e)
	}
}

func (c *compiler) checkExpressionValidList(dataType shared.List, expr parser.Expr) (err error) {
	var exprList parser.ExprList
	switch e := expr.(type) {
	case parser.ExprList:
		exprList = e
	default:
		return fmt.Errorf("expected %s got %T", dataType.String(), e)
	}

	for _, value := range exprList.Values {
		if err = c.checkExpressionValidDataType(dataType.ListType, value); err != nil {
			err = fmt.Errorf("failed to check if list type is valid data type: %w", err)
			return
		}
	}
	return nil
}

func (c *compiler) checkExpressionValidMap(dataType shared.Map, expr parser.Expr) (err error) {
	var exprTable parser.ExprTable
	switch e := expr.(type) {
	case parser.ExprTable:
		exprTable = e
	default:
		return fmt.Errorf("expected %s got %T", dataType.String(), e)
	}

	for _, pair := range exprTable.Pairs {
		if err = c.checkExpressionValidDataType(dataType.KeyType, pair.Key); err != nil {
			err = fmt.Errorf("failed to check if map key type is valid data type: %w", err)
			return
		}

		if err = c.checkExpressionValidDataType(dataType.ValueType, pair.Value); err != nil {
			err = fmt.Errorf("failed to check if map value type is valid data type: %w", err)
			return
		}
	}
	return nil
}

func (c *compiler) checkExpressionValidCustom(dataType shared.Custom, expr parser.Expr) (err error) {
	var exprTable parser.ExprTable
	switch e := expr.(type) {
	case parser.ExprTable:
		exprTable = e
	default:
		err = fmt.Errorf("expected %q got %T", dataType.Name, e)
		return
	}

	obj, ok := c.objects[dataType.Name]
	if !ok {
		err = fmt.Errorf("object %q not found", dataType.Name)
		return
	}

	for _, pair := range exprTable.Pairs {
		var keyName string
		switch kt := pair.Key.(type) {
		case parser.ExprVariable:
			keyName = kt.Name
		default:
			err = fmt.Errorf("field type %T not a label", kt)
			return
		}

		var keyFound bool
		var field parser.FieldTypePair
		for _, f := range obj.fields {
			if f.Field == keyName {
				keyFound = true
				field = f
			}
		}

		if !keyFound {
			err = fmt.Errorf("key %q not found in object %q", keyName, dataType.Name)
			return
		}

		if err = c.checkExpressionValidDataType(field.Type, pair.Value); err != nil {
			return err
		}
	}

	return nil
}

func (c *compiler) compileDataTypeZeroValue(dataType shared.DataType) (err error) {
	switch d := dataType.(type) {
	case shared.Custom:
		return c.compileObjectZeroValue(d)
	default:
		c.sb.WriteString(d.ZeroValue())
	}
	return nil
}

func (c *compiler) compileObjectZeroValue(dataType shared.Custom) (err error) {
	obj, ok := c.objects[dataType.Name]
	if !ok {
		err = fmt.Errorf("object %q does not exist", dataType.Name)
		return
	}

	fieldsLen := len(obj.fields)
	c.sb.WriteRune('{')
	for i, field := range obj.fields {
		c.sb.WriteRune('"')
		c.sb.WriteString(field.Field)
		c.sb.WriteRune('"')
		c.sb.WriteRune(':')
		if err = c.compileDataTypeZeroValue(field.Type); err != nil {
			return err
		}
		if i < fieldsLen-1 {
			c.sb.WriteRune(',')
		}
	}
	c.sb.WriteRune('}')
	return nil
}

func (c *compiler) compileExprIndex(expr parser.ExprIndex) (err error) {
	// Compile the left side (the container being indexed)
	if err = c.compileExpr(expr.Left); err != nil {
		err = fmt.Errorf("failed to compile left side of index: %w", err)
		return
	}

	// Check if the left expression is a variable to determine its type for index adjustment
	// For lists, we need to add 1 to the index since pixie is 0-indexed but Lua is 1-indexed
	needsIndexAdjustment := false

	// Determine if the container is a list by checking the variable type
	if varExpr, isVar := expr.Left.(parser.ExprVariable); isVar {
		if variable, exists := c.variables[varExpr.Name]; exists {
			// Check if the variable is of list type
			if _, isList := variable.dataType.(shared.List); isList {
				needsIndexAdjustment = true
			}
		}
	}

	c.sb.WriteRune('[')

	// If we need index adjustment and the index is a number literal, add 1 to it
	if needsIndexAdjustment {
		if _, isNum := expr.Index.(parser.ExprNumber); isNum {
			// For numeric indices, we add 1 to convert from pixie's 0-indexing to lua's 1-indexing
			c.sb.WriteString("(")
			if err = c.compileExpr(expr.Index); err != nil {
				err = fmt.Errorf("failed to compile index: %w", err)
				return
			}
			c.sb.WriteString(" + 1)")
		} else {
			// For non-numeric indices (like variables or expressions), wrap in parentheses and add 1
			c.sb.WriteString("(")
			if err = c.compileExpr(expr.Index); err != nil {
				err = fmt.Errorf("failed to compile index: %w", err)
				return
			}
			c.sb.WriteString(" + 1)")
		}
	} else {
		// For maps and other types, compile the index as-is
		if err = c.compileExpr(expr.Index); err != nil {
			err = fmt.Errorf("failed to compile index: %w", err)
			return
		}
	}

	c.sb.WriteRune(']')
	return nil
}

func (c *compiler) compileExprPropertyAccess(expr parser.ExprPropertyAccess) (err error) {
	// Compile the left side (the object being accessed)
	if err = c.compileExpr(expr.Left); err != nil {
		err = fmt.Errorf("failed to compile left side of property access: %w", err)
		return
	}

	// In Lua, object properties are accessed with dot notation like in Pixie
	c.sb.WriteRune('.')
	c.sb.WriteString(expr.Property)
	return nil
}
