package main

import (
	"fmt"
	"github.com/pingcap/parser"
	. "github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/format"
	_ "github.com/pingcap/tidb/types/parser_driver"
	"reflect"
	"strconv"
	"strings"
)

type printVisitor struct {
}

func (printVisitor) Enter(n Node) (node Node, skipChildren bool) {
	fmt.Printf("%#v\n", n)
	return n, false
}

func (printVisitor) Leave(n Node) (node Node, ok bool) {
	return n, true
}

// For test only.
func CleanNodeText(node Node) {
	var cleaner nodeTextCleaner
	node.Accept(&cleaner)
}

// nodeTextCleaner clean the text of a node and it's child node.
// For test only.
type nodeTextCleaner struct {
}

// Enter implements Visitor interface.
func (checker *nodeTextCleaner) Enter(in Node) (out Node, skipChildren bool) {
	in.SetText("")
	return in, false
}

// Leave implements Visitor interface.
func (checker *nodeTextCleaner) Leave(in Node) (out Node, ok bool) {
	return in, true
}

func Display(name string, x interface{}) {
	fmt.Printf("Display %s (%T):\n", name, x)
	display(name, reflect.ValueOf(x))
}

func formatAtom(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Invalid:
		return "invalid"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.String:
		return strconv.Quote(v.String())
	case reflect.Chan, reflect.Func, reflect.Ptr, reflect.Slice, reflect.Map:
		return v.Type().String() + "0x" + strconv.FormatUint(uint64(v.Pointer()), 16)
	default:
		return v.Type().String() + "value"
	}
}

func display(path string, v reflect.Value) {
	switch v.Kind() {
	case reflect.Invalid:
		fmt.Printf("%s = invalid\n", path)
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			display(fmt.Sprintf("%s[%d]", path, i), v.Index(i))
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			fieldPath := fmt.Sprintf("%s.%s", path, v.Type().Field(i).Name)
			display(fieldPath, v.Field(i))
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			display(fmt.Sprintf("%s[%s]", path, formatAtom(key)), v.MapIndex(key))
		}
	case reflect.Ptr:
		if v.IsNil() {
			fmt.Printf("%s = nil\n", path)
		} else {
			display(fmt.Sprintf("*%s", path), v.Elem())
		}
	case reflect.Interface:
		if v.IsNil() {
			fmt.Printf("%s = nil\n", path)
		} else {
			fmt.Printf("%s.type = %s\n", path, v.Elem().Type())
			display(path+".value", v.Elem())
		}
	default:
		fmt.Printf("%s = %s\n", path, formatAtom(v))

	}
}

func main() {
	parser := parser.New()
	stmt, err := parser.ParseOneStmt("CREATE TABLE bar (m INT) REPLACE SELECT n FROM foo;", "", "")
	if err != nil {
		fmt.Println(err.Error())
	}
	stmt1, err := parser.ParseOneStmt("CREATE TABLE foo (`name` CHAR(50) BINARY CHARACTER SET utf8)", "", "")
	if err != nil {
		fmt.Println(err.Error())
	}
	CleanNodeText(stmt)
	CleanNodeText(stmt1)
	Display("A", stmt)
	Display("B", stmt1)
	var sb strings.Builder
	stmt.(*CreateTableStmt).Cols[0].Restore(format.NewRestoreCtx(format.DefaultRestoreFlags, &sb))
	fmt.Println(sb.String())

	//stmt.Accept(printVisitor{})
	//fmt.Println()
	//fmt.Println()
	//fmt.Println()
	//fmt.Println()
	//stmt1.Accept(printVisitor{})

	fmt.Println(reflect.DeepEqual(stmt, stmt1))
}
