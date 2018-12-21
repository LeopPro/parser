package main

import (
	"fmt"
	"github.com/pingcap/parser"
	. "github.com/pingcap/parser/ast"
	_ "github.com/pingcap/tidb/types/parser_driver"
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

func main() {
	parser := parser.New()
	stmt, err := parser.ParseOneStmt("select `CONV`('a',16,2)", "", "")
	if err != nil {
		fmt.Println(err.Error())
	}
	stmt.Accept(printVisitor{})
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	stmt, err = parser.ParseOneStmt("select CONV('a',16,2)", "", "")
	if err != nil {
		fmt.Println(err.Error())
	}
	stmt.Accept(printVisitor{})
}
