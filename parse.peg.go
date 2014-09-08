package toml

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const end_symbol rune = 4

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleTOML
	ruleExpression
	rulenewline
	rulews
	rulewsnl
	rulecomment
	rulekeyval
	rulekey
	ruleval
	ruletable
	rulestdTable
	rulearrayTable
	ruletableKey
	ruletableKeySep
	ruleinteger
	ruleint
	rulefloat
	rulefrac
	ruleexp
	rulestring
	rulebasicString
	rulebasicChar
	ruleescaped
	rulebasicUnescaped
	ruleescape
	rulemlBasicString
	rulemlBasicBody
	ruleliteralString
	ruleliteralChar
	rulemlLiteralString
	rulemlLiteralBody
	rulemlLiteralChar
	rulehexdigit
	rulehexQuad
	ruleboolean
	ruledatetime
	ruledigit
	ruledigitDual
	ruledigitQuad
	rulearray
	rulearrayValues
	rulearraySep
	rulePegText
	ruleAction0
	ruleAction1
	ruleAction2
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8
	ruleAction9
	ruleAction10
	ruleAction11
	ruleAction12
	ruleAction13
	ruleAction14
	ruleAction15
	ruleAction16
	ruleAction17
	ruleAction18

	rulePre_
	rule_In_
	rule_Suf
)

var rul3s = [...]string{
	"Unknown",
	"TOML",
	"Expression",
	"newline",
	"ws",
	"wsnl",
	"comment",
	"keyval",
	"key",
	"val",
	"table",
	"stdTable",
	"arrayTable",
	"tableKey",
	"tableKeySep",
	"integer",
	"int",
	"float",
	"frac",
	"exp",
	"string",
	"basicString",
	"basicChar",
	"escaped",
	"basicUnescaped",
	"escape",
	"mlBasicString",
	"mlBasicBody",
	"literalString",
	"literalChar",
	"mlLiteralString",
	"mlLiteralBody",
	"mlLiteralChar",
	"hexdigit",
	"hexQuad",
	"boolean",
	"datetime",
	"digit",
	"digitDual",
	"digitQuad",
	"array",
	"arrayValues",
	"arraySep",
	"PegText",
	"Action0",
	"Action1",
	"Action2",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",
	"Action10",
	"Action11",
	"Action12",
	"Action13",
	"Action14",
	"Action15",
	"Action16",
	"Action17",
	"Action18",

	"Pre_",
	"_In_",
	"_Suf",
}

type tokenTree interface {
	Print()
	PrintSyntax()
	PrintSyntaxTree(buffer string)
	Add(rule pegRule, begin, end, next, depth int)
	Expand(index int) tokenTree
	Tokens() <-chan token32
	AST() *node32
	Error() []token32
	trim(length int)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(depth int, buffer string) {
	for node != nil {
		for c := 0; c < depth; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegRule], strconv.Quote(buffer[node.begin:node.end]))
		if node.up != nil {
			node.up.print(depth+1, buffer)
		}
		node = node.next
	}
}

func (ast *node32) Print(buffer string) {
	ast.print(0, buffer)
}

type element struct {
	node *node32
	down *element
}

/* ${@} bit structure for abstract syntax tree */
type token16 struct {
	pegRule
	begin, end, next int16
}

func (t *token16) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token16) isParentOf(u token16) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token16) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token16) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens16 struct {
	tree    []token16
	ordered [][]token16
}

func (t *tokens16) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens16) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens16) Order() [][]token16 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int16, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token16, len(depths)), make([]token16, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int16(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state16 struct {
	token16
	depths []int16
	leaf   bool
}

func (t *tokens16) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens16) PreOrder() (<-chan state16, [][]token16) {
	s, ordered := make(chan state16, 6), t.Order()
	go func() {
		var states [8]state16
		for i, _ := range states {
			states[i].depths = make([]int16, len(ordered))
		}
		depths, state, depth := make([]int16, len(ordered)), 0, 1
		write := func(t token16, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, int16(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token16 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token16{pegRule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token16{pegRule: rulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token16{pegRule: rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens16) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens16) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(buffer[token.begin:token.end]))
	}
}

func (t *tokens16) Add(rule pegRule, begin, end, depth, index int) {
	t.tree[index] = token16{pegRule: rule, begin: int16(begin), end: int16(end), next: int16(depth)}
}

func (t *tokens16) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens16) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	pegRule
	begin, end, next int32
}

func (t *token32) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens32) PreOrder() (<-chan state32, [][]token32) {
	s, ordered := make(chan state32, 6), t.Order()
	go func() {
		var states [8]state32
		for i, _ := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, int32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{pegRule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegRule: rulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{pegRule: rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(buffer[token.begin:token.end]))
	}
}

func (t *tokens32) Add(rule pegRule, begin, end, depth, index int) {
	t.tree[index] = token32{pegRule: rule, begin: int32(begin), end: int32(end), next: int32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

func (t *tokens16) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		for i, v := range tree {
			expanded[i] = v.getToken32()
		}
		return &tokens32{tree: expanded}
	}
	return nil
}

func (t *tokens32) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	return nil
}

type tomlParser struct {
	toml

	Buffer string
	buffer []rune
	rules  [63]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	tokenTree
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer string, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer[0:] {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p *tomlParser
}

func (e *parseError) Error() string {
	tokens, error := e.p.tokenTree.Error(), "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.Buffer, positions)
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf("parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n",
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			/*strconv.Quote(*/ e.p.Buffer[begin:end] /*)*/)
	}

	return error
}

func (p *tomlParser) PrintSyntaxTree() {
	p.tokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *tomlParser) Highlighter() {
	p.tokenTree.PrintSyntax()
}

func (p *tomlParser) Execute() {
	buffer, begin, end := p.Buffer, 0, 0
	for token := range p.tokenTree.Tokens() {
		switch token.pegRule {
		case rulePegText:
			begin, end = int(token.begin), int(token.end)
		case ruleAction0:
			p.AddLineCount(end - begin)
		case ruleAction1:
			p.AddLineCount(end - begin)
		case ruleAction2:
			p.AddKeyValue()
		case ruleAction3:
			p.SetKey(buffer[begin:end])
		case ruleAction4:
			p.SetTime(buffer[begin:end])
		case ruleAction5:
			p.SetFloat64(buffer[begin:end])
		case ruleAction6:
			p.SetInt64(buffer[begin:end])
		case ruleAction7:
			p.SetString()
		case ruleAction8:
			p.SetBool(buffer[begin:end])
		case ruleAction9:
			p.SetTable(buffer[begin:end])
		case ruleAction10:
			p.SetArrayTable(buffer[begin:end])
		case ruleAction11:
			p.SetBasicString(p.RuneSlice(buffer, begin, end))
		case ruleAction12:
			p.SetMultilineString()
		case ruleAction13:
			p.AddMultilineBasicBody(p.RuneSlice(buffer, begin, end))
		case ruleAction14:
			p.SetLiteralString(p.RuneSlice(buffer, begin, end))
		case ruleAction15:
			p.SetMultilineLiteralString(p.RuneSlice(buffer, begin, end))
		case ruleAction16:
			p.StartArray()
		case ruleAction17:
			p.EndArray()
		case ruleAction18:
			p.AddArrayVal()

		}
	}
}

func (p *tomlParser) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != end_symbol {
		p.buffer = append(p.buffer, end_symbol)
	}

	var tree tokenTree = &tokens16{tree: make([]token16, math.MaxInt16)}
	position, depth, tokenIndex, buffer, rules := 0, 0, 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokenTree = tree
		if matches {
			p.tokenTree.trim(tokenIndex)
			return nil
		}
		return &parseError{p}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule pegRule, begin int) {
		if t := tree.Expand(tokenIndex); t != nil {
			tree = t
		}
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
	}

	matchDot := func() bool {
		if buffer[position] != end_symbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	rules = [...]func() bool{
		nil,
		/* 0 TOML <- <(Expression (newline Expression)* newline? !.)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				if !rules[ruleExpression]() {
					goto l0
				}
			l2:
				{
					position3, tokenIndex3, depth3 := position, tokenIndex, depth
					if !rules[rulenewline]() {
						goto l3
					}
					if !rules[ruleExpression]() {
						goto l3
					}
					goto l2
				l3:
					position, tokenIndex, depth = position3, tokenIndex3, depth3
				}
				{
					position4, tokenIndex4, depth4 := position, tokenIndex, depth
					if !rules[rulenewline]() {
						goto l4
					}
					goto l5
				l4:
					position, tokenIndex, depth = position4, tokenIndex4, depth4
				}
			l5:
				{
					position6, tokenIndex6, depth6 := position, tokenIndex, depth
					if !matchDot() {
						goto l6
					}
					goto l0
				l6:
					position, tokenIndex, depth = position6, tokenIndex6, depth6
				}
				depth--
				add(ruleTOML, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 Expression <- <((ws table ws comment?) / (ws keyval ws comment?) / (ws comment?) / ws)> */
		func() bool {
			position7, tokenIndex7, depth7 := position, tokenIndex, depth
			{
				position8 := position
				depth++
				{
					position9, tokenIndex9, depth9 := position, tokenIndex, depth
					if !rules[rulews]() {
						goto l10
					}
					{
						position11 := position
						depth++
						{
							position12, tokenIndex12, depth12 := position, tokenIndex, depth
							{
								position14 := position
								depth++
								if buffer[position] != rune('[') {
									goto l13
								}
								position++
								if !rules[rulews]() {
									goto l13
								}
								{
									position15 := position
									depth++
									if !rules[ruletableKey]() {
										goto l13
									}
									depth--
									add(rulePegText, position15)
								}
								if !rules[rulews]() {
									goto l13
								}
								if buffer[position] != rune(']') {
									goto l13
								}
								position++
								{
									add(ruleAction9, position)
								}
								depth--
								add(rulestdTable, position14)
							}
							goto l12
						l13:
							position, tokenIndex, depth = position12, tokenIndex12, depth12
							{
								position17 := position
								depth++
								if buffer[position] != rune('[') {
									goto l10
								}
								position++
								if buffer[position] != rune('[') {
									goto l10
								}
								position++
								if !rules[rulews]() {
									goto l10
								}
								{
									position18 := position
									depth++
									if !rules[ruletableKey]() {
										goto l10
									}
									depth--
									add(rulePegText, position18)
								}
								if !rules[rulews]() {
									goto l10
								}
								if buffer[position] != rune(']') {
									goto l10
								}
								position++
								if buffer[position] != rune(']') {
									goto l10
								}
								position++
								{
									add(ruleAction10, position)
								}
								depth--
								add(rulearrayTable, position17)
							}
						}
					l12:
						depth--
						add(ruletable, position11)
					}
					if !rules[rulews]() {
						goto l10
					}
					{
						position20, tokenIndex20, depth20 := position, tokenIndex, depth
						if !rules[rulecomment]() {
							goto l20
						}
						goto l21
					l20:
						position, tokenIndex, depth = position20, tokenIndex20, depth20
					}
				l21:
					goto l9
				l10:
					position, tokenIndex, depth = position9, tokenIndex9, depth9
					if !rules[rulews]() {
						goto l22
					}
					{
						position23 := position
						depth++
						if !rules[rulekey]() {
							goto l22
						}
						if !rules[rulews]() {
							goto l22
						}
						if buffer[position] != rune('=') {
							goto l22
						}
						position++
						if !rules[rulews]() {
							goto l22
						}
						if !rules[ruleval]() {
							goto l22
						}
						{
							add(ruleAction2, position)
						}
						depth--
						add(rulekeyval, position23)
					}
					if !rules[rulews]() {
						goto l22
					}
					{
						position25, tokenIndex25, depth25 := position, tokenIndex, depth
						if !rules[rulecomment]() {
							goto l25
						}
						goto l26
					l25:
						position, tokenIndex, depth = position25, tokenIndex25, depth25
					}
				l26:
					goto l9
				l22:
					position, tokenIndex, depth = position9, tokenIndex9, depth9
					if !rules[rulews]() {
						goto l27
					}
					{
						position28, tokenIndex28, depth28 := position, tokenIndex, depth
						if !rules[rulecomment]() {
							goto l28
						}
						goto l29
					l28:
						position, tokenIndex, depth = position28, tokenIndex28, depth28
					}
				l29:
					goto l9
				l27:
					position, tokenIndex, depth = position9, tokenIndex9, depth9
					if !rules[rulews]() {
						goto l7
					}
				}
			l9:
				depth--
				add(ruleExpression, position8)
			}
			return true
		l7:
			position, tokenIndex, depth = position7, tokenIndex7, depth7
			return false
		},
		/* 2 newline <- <(<('\r' / '\n')+> Action0)> */
		func() bool {
			position30, tokenIndex30, depth30 := position, tokenIndex, depth
			{
				position31 := position
				depth++
				{
					position32 := position
					depth++
					{
						position35, tokenIndex35, depth35 := position, tokenIndex, depth
						if buffer[position] != rune('\r') {
							goto l36
						}
						position++
						goto l35
					l36:
						position, tokenIndex, depth = position35, tokenIndex35, depth35
						if buffer[position] != rune('\n') {
							goto l30
						}
						position++
					}
				l35:
				l33:
					{
						position34, tokenIndex34, depth34 := position, tokenIndex, depth
						{
							position37, tokenIndex37, depth37 := position, tokenIndex, depth
							if buffer[position] != rune('\r') {
								goto l38
							}
							position++
							goto l37
						l38:
							position, tokenIndex, depth = position37, tokenIndex37, depth37
							if buffer[position] != rune('\n') {
								goto l34
							}
							position++
						}
					l37:
						goto l33
					l34:
						position, tokenIndex, depth = position34, tokenIndex34, depth34
					}
					depth--
					add(rulePegText, position32)
				}
				{
					add(ruleAction0, position)
				}
				depth--
				add(rulenewline, position31)
			}
			return true
		l30:
			position, tokenIndex, depth = position30, tokenIndex30, depth30
			return false
		},
		/* 3 ws <- <(' ' / '\t')*> */
		func() bool {
			{
				position41 := position
				depth++
			l42:
				{
					position43, tokenIndex43, depth43 := position, tokenIndex, depth
					{
						position44, tokenIndex44, depth44 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l45
						}
						position++
						goto l44
					l45:
						position, tokenIndex, depth = position44, tokenIndex44, depth44
						if buffer[position] != rune('\t') {
							goto l43
						}
						position++
					}
				l44:
					goto l42
				l43:
					position, tokenIndex, depth = position43, tokenIndex43, depth43
				}
				depth--
				add(rulews, position41)
			}
			return true
		},
		/* 4 wsnl <- <((&('\t') '\t') | (&(' ') ' ') | (&('\n' | '\r') (<('\r' / '\n')> Action1)))*> */
		func() bool {
			{
				position47 := position
				depth++
			l48:
				{
					position49, tokenIndex49, depth49 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '\t':
							if buffer[position] != rune('\t') {
								goto l49
							}
							position++
							break
						case ' ':
							if buffer[position] != rune(' ') {
								goto l49
							}
							position++
							break
						default:
							{
								position51 := position
								depth++
								{
									position52, tokenIndex52, depth52 := position, tokenIndex, depth
									if buffer[position] != rune('\r') {
										goto l53
									}
									position++
									goto l52
								l53:
									position, tokenIndex, depth = position52, tokenIndex52, depth52
									if buffer[position] != rune('\n') {
										goto l49
									}
									position++
								}
							l52:
								depth--
								add(rulePegText, position51)
							}
							{
								add(ruleAction1, position)
							}
							break
						}
					}

					goto l48
				l49:
					position, tokenIndex, depth = position49, tokenIndex49, depth49
				}
				depth--
				add(rulewsnl, position47)
			}
			return true
		},
		/* 5 comment <- <('#' <('\t' / [ -􏿿])*>)> */
		func() bool {
			position55, tokenIndex55, depth55 := position, tokenIndex, depth
			{
				position56 := position
				depth++
				if buffer[position] != rune('#') {
					goto l55
				}
				position++
				{
					position57 := position
					depth++
				l58:
					{
						position59, tokenIndex59, depth59 := position, tokenIndex, depth
						{
							position60, tokenIndex60, depth60 := position, tokenIndex, depth
							if buffer[position] != rune('\t') {
								goto l61
							}
							position++
							goto l60
						l61:
							position, tokenIndex, depth = position60, tokenIndex60, depth60
							if c := buffer[position]; c < rune(' ') || c > rune('\U0010ffff') {
								goto l59
							}
							position++
						}
					l60:
						goto l58
					l59:
						position, tokenIndex, depth = position59, tokenIndex59, depth59
					}
					depth--
					add(rulePegText, position57)
				}
				depth--
				add(rulecomment, position56)
			}
			return true
		l55:
			position, tokenIndex, depth = position55, tokenIndex55, depth55
			return false
		},
		/* 6 keyval <- <(key ws '=' ws val Action2)> */
		nil,
		/* 7 key <- <(<((&('_') '_') | (&('-') '-') | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]))+> Action3)> */
		func() bool {
			position63, tokenIndex63, depth63 := position, tokenIndex, depth
			{
				position64 := position
				depth++
				{
					position65 := position
					depth++
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l63
							}
							position++
							break
						case '-':
							if buffer[position] != rune('-') {
								goto l63
							}
							position++
							break
						case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l63
							}
							position++
							break
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l63
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l63
							}
							position++
							break
						}
					}

				l66:
					{
						position67, tokenIndex67, depth67 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '_':
								if buffer[position] != rune('_') {
									goto l67
								}
								position++
								break
							case '-':
								if buffer[position] != rune('-') {
									goto l67
								}
								position++
								break
							case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l67
								}
								position++
								break
							case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l67
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l67
								}
								position++
								break
							}
						}

						goto l66
					l67:
						position, tokenIndex, depth = position67, tokenIndex67, depth67
					}
					depth--
					add(rulePegText, position65)
				}
				{
					add(ruleAction3, position)
				}
				depth--
				add(rulekey, position64)
			}
			return true
		l63:
			position, tokenIndex, depth = position63, tokenIndex63, depth63
			return false
		},
		/* 8 val <- <((<datetime> Action4) / (<float> Action5) / ((&('[') array) | (&('f' | 't') (<boolean> Action8)) | (&('"' | '\'') (string Action7)) | (&('-' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') (<integer> Action6))))> */
		func() bool {
			position71, tokenIndex71, depth71 := position, tokenIndex, depth
			{
				position72 := position
				depth++
				{
					position73, tokenIndex73, depth73 := position, tokenIndex, depth
					{
						position75 := position
						depth++
						{
							position76 := position
							depth++
							{
								position77 := position
								depth++
								if !rules[ruledigitDual]() {
									goto l74
								}
								if !rules[ruledigitDual]() {
									goto l74
								}
								depth--
								add(ruledigitQuad, position77)
							}
							if buffer[position] != rune('-') {
								goto l74
							}
							position++
							if !rules[ruledigitDual]() {
								goto l74
							}
							if buffer[position] != rune('-') {
								goto l74
							}
							position++
							if !rules[ruledigitDual]() {
								goto l74
							}
							if buffer[position] != rune('T') {
								goto l74
							}
							position++
							if !rules[ruledigitDual]() {
								goto l74
							}
							if buffer[position] != rune(':') {
								goto l74
							}
							position++
							if !rules[ruledigitDual]() {
								goto l74
							}
							if buffer[position] != rune(':') {
								goto l74
							}
							position++
							if !rules[ruledigitDual]() {
								goto l74
							}
							if buffer[position] != rune('Z') {
								goto l74
							}
							position++
							depth--
							add(ruledatetime, position76)
						}
						depth--
						add(rulePegText, position75)
					}
					{
						add(ruleAction4, position)
					}
					goto l73
				l74:
					position, tokenIndex, depth = position73, tokenIndex73, depth73
					{
						position80 := position
						depth++
						{
							position81 := position
							depth++
							{
								position82, tokenIndex82, depth82 := position, tokenIndex, depth
								if buffer[position] != rune('-') {
									goto l82
								}
								position++
								goto l83
							l82:
								position, tokenIndex, depth = position82, tokenIndex82, depth82
							}
						l83:
							if !rules[ruleint]() {
								goto l79
							}
							{
								position84 := position
								depth++
								if buffer[position] != rune('.') {
									goto l79
								}
								position++
								if !rules[ruledigit]() {
									goto l79
								}
							l85:
								{
									position86, tokenIndex86, depth86 := position, tokenIndex, depth
									if !rules[ruledigit]() {
										goto l86
									}
									goto l85
								l86:
									position, tokenIndex, depth = position86, tokenIndex86, depth86
								}
								depth--
								add(rulefrac, position84)
							}
							{
								position87, tokenIndex87, depth87 := position, tokenIndex, depth
								{
									position89 := position
									depth++
									{
										position90, tokenIndex90, depth90 := position, tokenIndex, depth
										if buffer[position] != rune('e') {
											goto l91
										}
										position++
										goto l90
									l91:
										position, tokenIndex, depth = position90, tokenIndex90, depth90
										if buffer[position] != rune('E') {
											goto l87
										}
										position++
									}
								l90:
									{
										position92, tokenIndex92, depth92 := position, tokenIndex, depth
										{
											position94, tokenIndex94, depth94 := position, tokenIndex, depth
											if buffer[position] != rune('-') {
												goto l95
											}
											position++
											goto l94
										l95:
											position, tokenIndex, depth = position94, tokenIndex94, depth94
											if buffer[position] != rune('+') {
												goto l92
											}
											position++
										}
									l94:
										goto l93
									l92:
										position, tokenIndex, depth = position92, tokenIndex92, depth92
									}
								l93:
									if !rules[ruledigit]() {
										goto l87
									}
								l96:
									{
										position97, tokenIndex97, depth97 := position, tokenIndex, depth
										if !rules[ruledigit]() {
											goto l97
										}
										goto l96
									l97:
										position, tokenIndex, depth = position97, tokenIndex97, depth97
									}
									depth--
									add(ruleexp, position89)
								}
								goto l88
							l87:
								position, tokenIndex, depth = position87, tokenIndex87, depth87
							}
						l88:
							depth--
							add(rulefloat, position81)
						}
						depth--
						add(rulePegText, position80)
					}
					{
						add(ruleAction5, position)
					}
					goto l73
				l79:
					position, tokenIndex, depth = position73, tokenIndex73, depth73
					{
						switch buffer[position] {
						case '[':
							{
								position100 := position
								depth++
								if buffer[position] != rune('[') {
									goto l71
								}
								position++
								{
									add(ruleAction16, position)
								}
								if !rules[rulewsnl]() {
									goto l71
								}
								{
									position102 := position
									depth++
								l103:
									{
										position104, tokenIndex104, depth104 := position, tokenIndex, depth
										if !rules[ruleval]() {
											goto l104
										}
										{
											add(ruleAction18, position)
										}
										{
											position106, tokenIndex106, depth106 := position, tokenIndex, depth
											{
												position108 := position
												depth++
												if !rules[rulews]() {
													goto l106
												}
												if buffer[position] != rune(',') {
													goto l106
												}
												position++
												if !rules[rulewsnl]() {
													goto l106
												}
												depth--
												add(rulearraySep, position108)
											}
											goto l107
										l106:
											position, tokenIndex, depth = position106, tokenIndex106, depth106
										}
									l107:
										{
											position109, tokenIndex109, depth109 := position, tokenIndex, depth
											{
												position111, tokenIndex111, depth111 := position, tokenIndex, depth
												if !rules[rulecomment]() {
													goto l111
												}
												goto l112
											l111:
												position, tokenIndex, depth = position111, tokenIndex111, depth111
											}
										l112:
											if !rules[rulenewline]() {
												goto l109
											}
											goto l110
										l109:
											position, tokenIndex, depth = position109, tokenIndex109, depth109
										}
									l110:
										goto l103
									l104:
										position, tokenIndex, depth = position104, tokenIndex104, depth104
									}
									depth--
									add(rulearrayValues, position102)
								}
								if !rules[rulewsnl]() {
									goto l71
								}
								if buffer[position] != rune(']') {
									goto l71
								}
								position++
								{
									add(ruleAction17, position)
								}
								depth--
								add(rulearray, position100)
							}
							break
						case 'f', 't':
							{
								position114 := position
								depth++
								{
									position115 := position
									depth++
									{
										position116, tokenIndex116, depth116 := position, tokenIndex, depth
										if buffer[position] != rune('t') {
											goto l117
										}
										position++
										if buffer[position] != rune('r') {
											goto l117
										}
										position++
										if buffer[position] != rune('u') {
											goto l117
										}
										position++
										if buffer[position] != rune('e') {
											goto l117
										}
										position++
										goto l116
									l117:
										position, tokenIndex, depth = position116, tokenIndex116, depth116
										if buffer[position] != rune('f') {
											goto l71
										}
										position++
										if buffer[position] != rune('a') {
											goto l71
										}
										position++
										if buffer[position] != rune('l') {
											goto l71
										}
										position++
										if buffer[position] != rune('s') {
											goto l71
										}
										position++
										if buffer[position] != rune('e') {
											goto l71
										}
										position++
									}
								l116:
									depth--
									add(ruleboolean, position115)
								}
								depth--
								add(rulePegText, position114)
							}
							{
								add(ruleAction8, position)
							}
							break
						case '"', '\'':
							{
								position119 := position
								depth++
								{
									position120, tokenIndex120, depth120 := position, tokenIndex, depth
									{
										position122 := position
										depth++
										if buffer[position] != rune('\'') {
											goto l121
										}
										position++
										if buffer[position] != rune('\'') {
											goto l121
										}
										position++
										if buffer[position] != rune('\'') {
											goto l121
										}
										position++
										{
											position123 := position
											depth++
											{
												position124 := position
												depth++
											l125:
												{
													position126, tokenIndex126, depth126 := position, tokenIndex, depth
													{
														position127, tokenIndex127, depth127 := position, tokenIndex, depth
														if buffer[position] != rune('\'') {
															goto l127
														}
														position++
														if buffer[position] != rune('\'') {
															goto l127
														}
														position++
														if buffer[position] != rune('\'') {
															goto l127
														}
														position++
														goto l126
													l127:
														position, tokenIndex, depth = position127, tokenIndex127, depth127
													}
													{
														position128, tokenIndex128, depth128 := position, tokenIndex, depth
														{
															position130 := position
															depth++
															{
																position131, tokenIndex131, depth131 := position, tokenIndex, depth
																if buffer[position] != rune('\t') {
																	goto l132
																}
																position++
																goto l131
															l132:
																position, tokenIndex, depth = position131, tokenIndex131, depth131
																if c := buffer[position]; c < rune(' ') || c > rune('\U0010ffff') {
																	goto l129
																}
																position++
															}
														l131:
															depth--
															add(rulemlLiteralChar, position130)
														}
														goto l128
													l129:
														position, tokenIndex, depth = position128, tokenIndex128, depth128
														if !rules[rulenewline]() {
															goto l126
														}
													}
												l128:
													goto l125
												l126:
													position, tokenIndex, depth = position126, tokenIndex126, depth126
												}
												depth--
												add(rulemlLiteralBody, position124)
											}
											depth--
											add(rulePegText, position123)
										}
										if buffer[position] != rune('\'') {
											goto l121
										}
										position++
										if buffer[position] != rune('\'') {
											goto l121
										}
										position++
										if buffer[position] != rune('\'') {
											goto l121
										}
										position++
										{
											add(ruleAction15, position)
										}
										depth--
										add(rulemlLiteralString, position122)
									}
									goto l120
								l121:
									position, tokenIndex, depth = position120, tokenIndex120, depth120
									{
										position135 := position
										depth++
										if buffer[position] != rune('\'') {
											goto l134
										}
										position++
										{
											position136 := position
											depth++
										l137:
											{
												position138, tokenIndex138, depth138 := position, tokenIndex, depth
												{
													position139 := position
													depth++
													{
														switch buffer[position] {
														case '\t':
															if buffer[position] != rune('\t') {
																goto l138
															}
															position++
															break
														case ' ', '!', '"', '#', '$', '%', '&':
															if c := buffer[position]; c < rune(' ') || c > rune('&') {
																goto l138
															}
															position++
															break
														default:
															if c := buffer[position]; c < rune('(') || c > rune('\U0010ffff') {
																goto l138
															}
															position++
															break
														}
													}

													depth--
													add(ruleliteralChar, position139)
												}
												goto l137
											l138:
												position, tokenIndex, depth = position138, tokenIndex138, depth138
											}
											depth--
											add(rulePegText, position136)
										}
										if buffer[position] != rune('\'') {
											goto l134
										}
										position++
										{
											add(ruleAction14, position)
										}
										depth--
										add(ruleliteralString, position135)
									}
									goto l120
								l134:
									position, tokenIndex, depth = position120, tokenIndex120, depth120
									{
										position143 := position
										depth++
										if buffer[position] != rune('"') {
											goto l142
										}
										position++
										if buffer[position] != rune('"') {
											goto l142
										}
										position++
										if buffer[position] != rune('"') {
											goto l142
										}
										position++
										{
											position144 := position
											depth++
										l145:
											{
												position146, tokenIndex146, depth146 := position, tokenIndex, depth
												{
													position147, tokenIndex147, depth147 := position, tokenIndex, depth
													{
														position149 := position
														depth++
														{
															position150, tokenIndex150, depth150 := position, tokenIndex, depth
															if !rules[rulebasicChar]() {
																goto l151
															}
															goto l150
														l151:
															position, tokenIndex, depth = position150, tokenIndex150, depth150
															if !rules[rulenewline]() {
																goto l148
															}
														}
													l150:
														depth--
														add(rulePegText, position149)
													}
													{
														add(ruleAction13, position)
													}
													goto l147
												l148:
													position, tokenIndex, depth = position147, tokenIndex147, depth147
													if !rules[ruleescape]() {
														goto l146
													}
													if !rules[rulenewline]() {
														goto l146
													}
													if !rules[rulewsnl]() {
														goto l146
													}
												}
											l147:
												goto l145
											l146:
												position, tokenIndex, depth = position146, tokenIndex146, depth146
											}
											depth--
											add(rulemlBasicBody, position144)
										}
										if buffer[position] != rune('"') {
											goto l142
										}
										position++
										if buffer[position] != rune('"') {
											goto l142
										}
										position++
										if buffer[position] != rune('"') {
											goto l142
										}
										position++
										{
											add(ruleAction12, position)
										}
										depth--
										add(rulemlBasicString, position143)
									}
									goto l120
								l142:
									position, tokenIndex, depth = position120, tokenIndex120, depth120
									{
										position154 := position
										depth++
										{
											position155 := position
											depth++
											if buffer[position] != rune('"') {
												goto l71
											}
											position++
										l156:
											{
												position157, tokenIndex157, depth157 := position, tokenIndex, depth
												if !rules[rulebasicChar]() {
													goto l157
												}
												goto l156
											l157:
												position, tokenIndex, depth = position157, tokenIndex157, depth157
											}
											if buffer[position] != rune('"') {
												goto l71
											}
											position++
											depth--
											add(rulePegText, position155)
										}
										{
											add(ruleAction11, position)
										}
										depth--
										add(rulebasicString, position154)
									}
								}
							l120:
								depth--
								add(rulestring, position119)
							}
							{
								add(ruleAction7, position)
							}
							break
						default:
							{
								position160 := position
								depth++
								{
									position161 := position
									depth++
									{
										position162, tokenIndex162, depth162 := position, tokenIndex, depth
										if buffer[position] != rune('-') {
											goto l162
										}
										position++
										goto l163
									l162:
										position, tokenIndex, depth = position162, tokenIndex162, depth162
									}
								l163:
									if !rules[ruleint]() {
										goto l71
									}
									depth--
									add(ruleinteger, position161)
								}
								depth--
								add(rulePegText, position160)
							}
							{
								add(ruleAction6, position)
							}
							break
						}
					}

				}
			l73:
				depth--
				add(ruleval, position72)
			}
			return true
		l71:
			position, tokenIndex, depth = position71, tokenIndex71, depth71
			return false
		},
		/* 9 table <- <(stdTable / arrayTable)> */
		nil,
		/* 10 stdTable <- <('[' ws <tableKey> ws ']' Action9)> */
		nil,
		/* 11 arrayTable <- <('[' '[' ws <tableKey> ws (']' ']') Action10)> */
		nil,
		/* 12 tableKey <- <(key (tableKeySep key)*)> */
		func() bool {
			position168, tokenIndex168, depth168 := position, tokenIndex, depth
			{
				position169 := position
				depth++
				if !rules[rulekey]() {
					goto l168
				}
			l170:
				{
					position171, tokenIndex171, depth171 := position, tokenIndex, depth
					{
						position172 := position
						depth++
						if !rules[rulews]() {
							goto l171
						}
						if buffer[position] != rune('.') {
							goto l171
						}
						position++
						if !rules[rulews]() {
							goto l171
						}
						depth--
						add(ruletableKeySep, position172)
					}
					if !rules[rulekey]() {
						goto l171
					}
					goto l170
				l171:
					position, tokenIndex, depth = position171, tokenIndex171, depth171
				}
				depth--
				add(ruletableKey, position169)
			}
			return true
		l168:
			position, tokenIndex, depth = position168, tokenIndex168, depth168
			return false
		},
		/* 13 tableKeySep <- <(ws '.' ws)> */
		nil,
		/* 14 integer <- <('-'? int)> */
		nil,
		/* 15 int <- <('0' / ([1-9] digit*))> */
		func() bool {
			position175, tokenIndex175, depth175 := position, tokenIndex, depth
			{
				position176 := position
				depth++
				{
					position177, tokenIndex177, depth177 := position, tokenIndex, depth
					if buffer[position] != rune('0') {
						goto l178
					}
					position++
					goto l177
				l178:
					position, tokenIndex, depth = position177, tokenIndex177, depth177
					if c := buffer[position]; c < rune('1') || c > rune('9') {
						goto l175
					}
					position++
				l179:
					{
						position180, tokenIndex180, depth180 := position, tokenIndex, depth
						if !rules[ruledigit]() {
							goto l180
						}
						goto l179
					l180:
						position, tokenIndex, depth = position180, tokenIndex180, depth180
					}
				}
			l177:
				depth--
				add(ruleint, position176)
			}
			return true
		l175:
			position, tokenIndex, depth = position175, tokenIndex175, depth175
			return false
		},
		/* 16 float <- <('-'? int frac exp?)> */
		nil,
		/* 17 frac <- <('.' digit+)> */
		nil,
		/* 18 exp <- <(('e' / 'E') ('-' / '+')? digit+)> */
		nil,
		/* 19 string <- <(mlLiteralString / literalString / mlBasicString / basicString)> */
		nil,
		/* 20 basicString <- <(<('"' basicChar* '"')> Action11)> */
		nil,
		/* 21 basicChar <- <(basicUnescaped / escaped)> */
		func() bool {
			position186, tokenIndex186, depth186 := position, tokenIndex, depth
			{
				position187 := position
				depth++
				{
					position188, tokenIndex188, depth188 := position, tokenIndex, depth
					{
						position190 := position
						depth++
						{
							switch buffer[position] {
							case ' ', '!':
								if c := buffer[position]; c < rune(' ') || c > rune('!') {
									goto l189
								}
								position++
								break
							case '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ':', ';', '<', '=', '>', '?', '@', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '[':
								if c := buffer[position]; c < rune('#') || c > rune('[') {
									goto l189
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune(']') || c > rune('\U0010ffff') {
									goto l189
								}
								position++
								break
							}
						}

						depth--
						add(rulebasicUnescaped, position190)
					}
					goto l188
				l189:
					position, tokenIndex, depth = position188, tokenIndex188, depth188
					{
						position192 := position
						depth++
						if !rules[ruleescape]() {
							goto l186
						}
						{
							switch buffer[position] {
							case 'U':
								if buffer[position] != rune('U') {
									goto l186
								}
								position++
								if !rules[rulehexQuad]() {
									goto l186
								}
								if !rules[rulehexQuad]() {
									goto l186
								}
								break
							case 'u':
								if buffer[position] != rune('u') {
									goto l186
								}
								position++
								if !rules[rulehexQuad]() {
									goto l186
								}
								break
							case '\\':
								if buffer[position] != rune('\\') {
									goto l186
								}
								position++
								break
							case '/':
								if buffer[position] != rune('/') {
									goto l186
								}
								position++
								break
							case '"':
								if buffer[position] != rune('"') {
									goto l186
								}
								position++
								break
							case 'r':
								if buffer[position] != rune('r') {
									goto l186
								}
								position++
								break
							case 'f':
								if buffer[position] != rune('f') {
									goto l186
								}
								position++
								break
							case 'n':
								if buffer[position] != rune('n') {
									goto l186
								}
								position++
								break
							case 't':
								if buffer[position] != rune('t') {
									goto l186
								}
								position++
								break
							default:
								if buffer[position] != rune('b') {
									goto l186
								}
								position++
								break
							}
						}

						depth--
						add(ruleescaped, position192)
					}
				}
			l188:
				depth--
				add(rulebasicChar, position187)
			}
			return true
		l186:
			position, tokenIndex, depth = position186, tokenIndex186, depth186
			return false
		},
		/* 22 escaped <- <(escape ((&('U') ('U' hexQuad hexQuad)) | (&('u') ('u' hexQuad)) | (&('\\') '\\') | (&('/') '/') | (&('"') '"') | (&('r') 'r') | (&('f') 'f') | (&('n') 'n') | (&('t') 't') | (&('b') 'b')))> */
		nil,
		/* 23 basicUnescaped <- <((&(' ' | '!') [ -!]) | (&('#' | '$' | '%' | '&' | '\'' | '(' | ')' | '*' | '+' | ',' | '-' | '.' | '/' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | ':' | ';' | '<' | '=' | '>' | '?' | '@' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '[') [#-[]) | (&(']' | '^' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z' | '{' | '|' | '}' | '~' | '\u007f' | '\u0080' | '\u0081' | '\u0082' | '\u0083' | '\u0084' | '\u0085' | '\u0086' | '\u0087' | '\u0088' | '\u0089' | '\u008a' | '\u008b' | '\u008c' | '\u008d' | '\u008e' | '\u008f' | '\u0090' | '\u0091' | '\u0092' | '\u0093' | '\u0094' | '\u0095' | '\u0096' | '\u0097' | '\u0098' | '\u0099' | '\u009a' | '\u009b' | '\u009c' | '\u009d' | '\u009e' | '\u009f' | '\u00a0' | '¡' | '¢' | '£' | '¤' | '¥' | '¦' | '§' | '¨' | '©' | 'ª' | '«' | '¬' | '\u00ad' | '®' | '¯' | '°' | '±' | '²' | '³' | '´' | 'µ' | '¶' | '·' | '¸' | '¹' | 'º' | '»' | '¼' | '½' | '¾' | '¿' | 'À' | 'Á' | 'Â' | 'Ã' | 'Ä' | 'Å' | 'Æ' | 'Ç' | 'È' | 'É' | 'Ê' | 'Ë' | 'Ì' | 'Í' | 'Î' | 'Ï' | 'Ð' | 'Ñ' | 'Ò' | 'Ó' | 'Ô' | 'Õ' | 'Ö' | '×' | 'Ø' | 'Ù' | 'Ú' | 'Û' | 'Ü' | 'Ý' | 'Þ' | 'ß' | 'à' | 'á' | 'â' | 'ã' | 'ä' | 'å' | 'æ' | 'ç' | 'è' | 'é' | 'ê' | 'ë' | 'ì' | 'í' | 'î' | 'ï' | 'ð' | 'ñ' | 'ò' | 'ó' | 'ô') []-􏿿]))> */
		nil,
		/* 24 escape <- <'\\'> */
		func() bool {
			position196, tokenIndex196, depth196 := position, tokenIndex, depth
			{
				position197 := position
				depth++
				if buffer[position] != rune('\\') {
					goto l196
				}
				position++
				depth--
				add(ruleescape, position197)
			}
			return true
		l196:
			position, tokenIndex, depth = position196, tokenIndex196, depth196
			return false
		},
		/* 25 mlBasicString <- <('"' '"' '"' mlBasicBody ('"' '"' '"') Action12)> */
		nil,
		/* 26 mlBasicBody <- <((<(basicChar / newline)> Action13) / (escape newline wsnl))*> */
		nil,
		/* 27 literalString <- <('\'' <literalChar*> '\'' Action14)> */
		nil,
		/* 28 literalChar <- <((&('\t') '\t') | (&(' ' | '!' | '"' | '#' | '$' | '%' | '&') [ -&]) | (&('(' | ')' | '*' | '+' | ',' | '-' | '.' | '/' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | ':' | ';' | '<' | '=' | '>' | '?' | '@' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '[' | '\\' | ']' | '^' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z' | '{' | '|' | '}' | '~' | '\u007f' | '\u0080' | '\u0081' | '\u0082' | '\u0083' | '\u0084' | '\u0085' | '\u0086' | '\u0087' | '\u0088' | '\u0089' | '\u008a' | '\u008b' | '\u008c' | '\u008d' | '\u008e' | '\u008f' | '\u0090' | '\u0091' | '\u0092' | '\u0093' | '\u0094' | '\u0095' | '\u0096' | '\u0097' | '\u0098' | '\u0099' | '\u009a' | '\u009b' | '\u009c' | '\u009d' | '\u009e' | '\u009f' | '\u00a0' | '¡' | '¢' | '£' | '¤' | '¥' | '¦' | '§' | '¨' | '©' | 'ª' | '«' | '¬' | '\u00ad' | '®' | '¯' | '°' | '±' | '²' | '³' | '´' | 'µ' | '¶' | '·' | '¸' | '¹' | 'º' | '»' | '¼' | '½' | '¾' | '¿' | 'À' | 'Á' | 'Â' | 'Ã' | 'Ä' | 'Å' | 'Æ' | 'Ç' | 'È' | 'É' | 'Ê' | 'Ë' | 'Ì' | 'Í' | 'Î' | 'Ï' | 'Ð' | 'Ñ' | 'Ò' | 'Ó' | 'Ô' | 'Õ' | 'Ö' | '×' | 'Ø' | 'Ù' | 'Ú' | 'Û' | 'Ü' | 'Ý' | 'Þ' | 'ß' | 'à' | 'á' | 'â' | 'ã' | 'ä' | 'å' | 'æ' | 'ç' | 'è' | 'é' | 'ê' | 'ë' | 'ì' | 'í' | 'î' | 'ï' | 'ð' | 'ñ' | 'ò' | 'ó' | 'ô') [(-􏿿]))> */
		nil,
		/* 29 mlLiteralString <- <('\'' '\'' '\'' <mlLiteralBody> ('\'' '\'' '\'') Action15)> */
		nil,
		/* 30 mlLiteralBody <- <(!('\'' '\'' '\'') (mlLiteralChar / newline))*> */
		nil,
		/* 31 mlLiteralChar <- <('\t' / [ -􏿿])> */
		nil,
		/* 32 hexdigit <- <((&('a' | 'b' | 'c' | 'd' | 'e' | 'f') [a-f]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F') [A-F]) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]))> */
		func() bool {
			position205, tokenIndex205, depth205 := position, tokenIndex, depth
			{
				position206 := position
				depth++
				{
					switch buffer[position] {
					case 'a', 'b', 'c', 'd', 'e', 'f':
						if c := buffer[position]; c < rune('a') || c > rune('f') {
							goto l205
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F':
						if c := buffer[position]; c < rune('A') || c > rune('F') {
							goto l205
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l205
						}
						position++
						break
					}
				}

				depth--
				add(rulehexdigit, position206)
			}
			return true
		l205:
			position, tokenIndex, depth = position205, tokenIndex205, depth205
			return false
		},
		/* 33 hexQuad <- <(hexdigit hexdigit hexdigit hexdigit)> */
		func() bool {
			position208, tokenIndex208, depth208 := position, tokenIndex, depth
			{
				position209 := position
				depth++
				if !rules[rulehexdigit]() {
					goto l208
				}
				if !rules[rulehexdigit]() {
					goto l208
				}
				if !rules[rulehexdigit]() {
					goto l208
				}
				if !rules[rulehexdigit]() {
					goto l208
				}
				depth--
				add(rulehexQuad, position209)
			}
			return true
		l208:
			position, tokenIndex, depth = position208, tokenIndex208, depth208
			return false
		},
		/* 34 boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		nil,
		/* 35 datetime <- <(digitQuad '-' digitDual '-' digitDual 'T' digitDual ':' digitDual ':' digitDual 'Z')> */
		nil,
		/* 36 digit <- <[0-9]> */
		func() bool {
			position212, tokenIndex212, depth212 := position, tokenIndex, depth
			{
				position213 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l212
				}
				position++
				depth--
				add(ruledigit, position213)
			}
			return true
		l212:
			position, tokenIndex, depth = position212, tokenIndex212, depth212
			return false
		},
		/* 37 digitDual <- <(digit digit)> */
		func() bool {
			position214, tokenIndex214, depth214 := position, tokenIndex, depth
			{
				position215 := position
				depth++
				if !rules[ruledigit]() {
					goto l214
				}
				if !rules[ruledigit]() {
					goto l214
				}
				depth--
				add(ruledigitDual, position215)
			}
			return true
		l214:
			position, tokenIndex, depth = position214, tokenIndex214, depth214
			return false
		},
		/* 38 digitQuad <- <(digitDual digitDual)> */
		nil,
		/* 39 array <- <('[' Action16 wsnl arrayValues wsnl ']' Action17)> */
		nil,
		/* 40 arrayValues <- <(val Action18 arraySep? (comment? newline)?)*> */
		nil,
		/* 41 arraySep <- <(ws ',' wsnl)> */
		nil,
		nil,
		/* 44 Action0 <- <{ p.AddLineCount(end - begin) }> */
		nil,
		/* 45 Action1 <- <{ p.AddLineCount(end - begin) }> */
		nil,
		/* 46 Action2 <- <{ p.AddKeyValue() }> */
		nil,
		/* 47 Action3 <- <{ p.SetKey(buffer[begin:end]) }> */
		nil,
		/* 48 Action4 <- <{ p.SetTime(buffer[begin:end]) }> */
		nil,
		/* 49 Action5 <- <{ p.SetFloat64(buffer[begin:end]) }> */
		nil,
		/* 50 Action6 <- <{ p.SetInt64(buffer[begin:end]) }> */
		nil,
		/* 51 Action7 <- <{ p.SetString() }> */
		nil,
		/* 52 Action8 <- <{ p.SetBool(buffer[begin:end]) }> */
		nil,
		/* 53 Action9 <- <{ p.SetTable(buffer[begin:end]) }> */
		nil,
		/* 54 Action10 <- <{ p.SetArrayTable(buffer[begin:end]) }> */
		nil,
		/* 55 Action11 <- <{ p.SetBasicString(p.RuneSlice(buffer, begin, end)) }> */
		nil,
		/* 56 Action12 <- <{ p.SetMultilineString() }> */
		nil,
		/* 57 Action13 <- <{ p.AddMultilineBasicBody(p.RuneSlice(buffer, begin, end)) }> */
		nil,
		/* 58 Action14 <- <{ p.SetLiteralString(p.RuneSlice(buffer, begin, end)) }> */
		nil,
		/* 59 Action15 <- <{ p.SetMultilineLiteralString(p.RuneSlice(buffer, begin, end)) }> */
		nil,
		/* 60 Action16 <- <{ p.StartArray() }> */
		nil,
		/* 61 Action17 <- <{ p.EndArray() }> */
		nil,
		/* 62 Action18 <- <{ p.AddArrayVal() }> */
		nil,
	}
	p.rules = rules
}
