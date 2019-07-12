package sql_generator

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/pingcap/parser/compatibility_reporter/yacc_parser"
)

type node interface {
	walk() bool
	materialize(writer io.StringWriter) error
}

type literalNode struct {
	value string
}

func (ln *literalNode) walk() bool {
	return true
}

func (ln *literalNode) materialize(writer io.StringWriter) error {
	if len(ln.value) != 0 {
		_, err := writer.WriteString(ln.value)
		if err != nil {
			return err
		}
		_, err = writer.WriteString(" ")
		if err != nil {
			return err
		}
	}
	return nil
}

type terminator struct {
}

func (t *terminator) walk() bool {
	panic("unreachable")
}

func (t *terminator) materialize(writer io.StringWriter) error {
	panic("unreachable")
}

type expressionNode struct {
	items []node
}

func (en *expressionNode) materialize(writer io.StringWriter) error {
	for _, item := range en.items {
		err := item.materialize(writer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (en *expressionNode) walk() (carry bool) {
	previousCarry := true
	for i := len(en.items) - 1; i >= 0 && previousCarry; i-- {
		previousCarry = en.items[i].walk()
	}
	return previousCarry
}

func (en *expressionNode) existTerminator() bool {
	for _, item := range en.items {
		if _, ok := item.(*terminator); ok {
			return true
		}
	}
	for _, item := range en.items {
		if pNode, ok := item.(*productionNode); ok {
			pNode.pruneTerminator()
		}
	}
	return false
}

type productionNode struct {
	name      string
	exprs     []*expressionNode
	fathers   []string
	walkIndex int
	pruned    bool
}

func (pn *productionNode) walk() (carry bool) {
	if pn.exprs[pn.walkIndex].walk() {
		pn.walkIndex += 1
		if pn.walkIndex >= len(pn.exprs) {
			pn.walkIndex = 0
			carry = true
		}
	}
	return
}

func (pn *productionNode) materialize(writer io.StringWriter) error {
	return pn.exprs[pn.walkIndex].materialize(writer)
}

func (pn *productionNode) pruneTerminator() {
	if pn.pruned {
		return
	}
	var newExprs []*expressionNode
	for _, expr := range pn.exprs {
		if !expr.existTerminator() {
			newExprs = append(newExprs, expr)
		}
	}
	pn.exprs = newExprs
	pn.pruned = true
}

func newProductionNode(production yacc_parser.Production, fathers []string) *productionNode {
	return &productionNode{
		name:    production.Head,
		exprs:   make([]*expressionNode, len(production.Alter)),
		fathers: fathers,
	}
}

func newExpressionNode(seq yacc_parser.Seq) *expressionNode {
	return &expressionNode{
		items: make([]node, len(seq.Items)),
	}
}

var productionMap = make(map[string]yacc_parser.Production)

func buildTree(productionName string, fathers []string) node {
	sumFather := 0
	for _, father := range fathers {
		if father == productionName {
			sumFather += 1
		}
	}
	if sumFather >= 2 {
		return &terminator{}
	}
	production, exist := productionMap[productionName]
	if !exist {
		panic(fmt.Sprintf("Production '%s' not found", productionName))
	}
	root := newProductionNode(production, fathers)
	for i, seq := range production.Alter {
		root.exprs[i] = newExpressionNode(seq)
		for j, item := range seq.Items {
			if strings.HasPrefix(item, "'") && strings.HasSuffix(item, "'") {
				root.exprs[i].items[j] = &literalNode{value: strings.Trim(item, "'")}
			} else {
				root.exprs[i].items[j] = buildTree(item, append(fathers, productionName))
			}
		}
	}
	return root
}

type SQLIterator struct {
	root             *productionNode
	alreadyPointNext bool
	noNext           bool
}

func (i *SQLIterator) HasNext() bool {
	if !i.alreadyPointNext {
		i.noNext = i.root.walk()
		i.alreadyPointNext = true
	}
	return !i.noNext
}

func (i *SQLIterator) Next() string {
	if !i.HasNext() {
		panic("there isn't next item in this sql iterator")
	}
	i.alreadyPointNext = false
	stringBuffer := bytes.NewBuffer([]byte{})
	err := i.root.materialize(stringBuffer)
	if err != nil {
		panic("buffer write failure" + err.Error())
	}
	return stringBuffer.String()
}

func GenerateSQL(productions []yacc_parser.Production, productionName string) *SQLIterator {
	for _, production := range productions {
		if _, exist := productionMap[production.Head]; exist {
			panic(fmt.Sprintf("Production '%s' duplicate definitions", production.Head))
		}
		productionMap[production.Head] = production
	}
	pNode := buildTree(productionName, nil).(*productionNode)
	pNode.pruneTerminator()
	return &SQLIterator{
		root:             pNode,
		alreadyPointNext: true,
		noNext:           false,
	}
}
