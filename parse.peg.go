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
	ruleAction19

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
	"Action19",

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
	rules  [64]func() bool
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
			p.SetTableString(begin, end)
		case ruleAction1:
			p.AddLineCount(end - begin)
		case ruleAction2:
			p.AddLineCount(end - begin)
		case ruleAction3:
			p.AddKeyValue()
		case ruleAction4:
			p.SetKey(buffer[begin:end])
		case ruleAction5:
			p.SetTime(begin, end)
		case ruleAction6:
			p.SetFloat64(begin, end)
		case ruleAction7:
			p.SetInt64(begin, end)
		case ruleAction8:
			p.SetString(begin, end)
		case ruleAction9:
			p.SetBool(begin, end)
		case ruleAction10:
			p.SetArray(begin, end)
		case ruleAction11:
			p.SetTable(buffer[begin:end])
		case ruleAction12:
			p.SetArrayTable(buffer[begin:end])
		case ruleAction13:
			p.SetBasicString(p.RuneSlice(buffer, begin, end))
		case ruleAction14:
			p.SetMultilineString()
		case ruleAction15:
			p.AddMultilineBasicBody(p.RuneSlice(buffer, begin, end))
		case ruleAction16:
			p.SetLiteralString(p.RuneSlice(buffer, begin, end))
		case ruleAction17:
			p.SetMultilineLiteralString(p.RuneSlice(buffer, begin, end))
		case ruleAction18:
			p.StartArray()
		case ruleAction19:
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
		/* 1 Expression <- <((<(ws table ws comment? (wsnl keyval ws comment?)*)> Action0) / (ws keyval ws comment?) / (ws comment?) / ws)> */
		func() bool {
			position7, tokenIndex7, depth7 := position, tokenIndex, depth
			{
				position8 := position
				depth++
				{
					position9, tokenIndex9, depth9 := position, tokenIndex, depth
					{
						position11 := position
						depth++
						if !rules[rulews]() {
							goto l10
						}
						{
							position12 := position
							depth++
							{
								position13, tokenIndex13, depth13 := position, tokenIndex, depth
								{
									position15 := position
									depth++
									if buffer[position] != rune('[') {
										goto l14
									}
									position++
									if !rules[rulews]() {
										goto l14
									}
									{
										position16 := position
										depth++
										if !rules[ruletableKey]() {
											goto l14
										}
										depth--
										add(rulePegText, position16)
									}
									if !rules[rulews]() {
										goto l14
									}
									if buffer[position] != rune(']') {
										goto l14
									}
									position++
									{
										add(ruleAction11, position)
									}
									depth--
									add(rulestdTable, position15)
								}
								goto l13
							l14:
								position, tokenIndex, depth = position13, tokenIndex13, depth13
								{
									position18 := position
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
										position19 := position
										depth++
										if !rules[ruletableKey]() {
											goto l10
										}
										depth--
										add(rulePegText, position19)
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
										add(ruleAction12, position)
									}
									depth--
									add(rulearrayTable, position18)
								}
							}
						l13:
							depth--
							add(ruletable, position12)
						}
						if !rules[rulews]() {
							goto l10
						}
						{
							position21, tokenIndex21, depth21 := position, tokenIndex, depth
							if !rules[rulecomment]() {
								goto l21
							}
							goto l22
						l21:
							position, tokenIndex, depth = position21, tokenIndex21, depth21
						}
					l22:
					l23:
						{
							position24, tokenIndex24, depth24 := position, tokenIndex, depth
							if !rules[rulewsnl]() {
								goto l24
							}
							if !rules[rulekeyval]() {
								goto l24
							}
							if !rules[rulews]() {
								goto l24
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
							goto l23
						l24:
							position, tokenIndex, depth = position24, tokenIndex24, depth24
						}
						depth--
						add(rulePegText, position11)
					}
					{
						add(ruleAction0, position)
					}
					goto l9
				l10:
					position, tokenIndex, depth = position9, tokenIndex9, depth9
					if !rules[rulews]() {
						goto l28
					}
					if !rules[rulekeyval]() {
						goto l28
					}
					if !rules[rulews]() {
						goto l28
					}
					{
						position29, tokenIndex29, depth29 := position, tokenIndex, depth
						if !rules[rulecomment]() {
							goto l29
						}
						goto l30
					l29:
						position, tokenIndex, depth = position29, tokenIndex29, depth29
					}
				l30:
					goto l9
				l28:
					position, tokenIndex, depth = position9, tokenIndex9, depth9
					if !rules[rulews]() {
						goto l31
					}
					{
						position32, tokenIndex32, depth32 := position, tokenIndex, depth
						if !rules[rulecomment]() {
							goto l32
						}
						goto l33
					l32:
						position, tokenIndex, depth = position32, tokenIndex32, depth32
					}
				l33:
					goto l9
				l31:
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
		/* 2 newline <- <(<('\r' / '\n')+> Action1)> */
		func() bool {
			position34, tokenIndex34, depth34 := position, tokenIndex, depth
			{
				position35 := position
				depth++
				{
					position36 := position
					depth++
					{
						position39, tokenIndex39, depth39 := position, tokenIndex, depth
						if buffer[position] != rune('\r') {
							goto l40
						}
						position++
						goto l39
					l40:
						position, tokenIndex, depth = position39, tokenIndex39, depth39
						if buffer[position] != rune('\n') {
							goto l34
						}
						position++
					}
				l39:
				l37:
					{
						position38, tokenIndex38, depth38 := position, tokenIndex, depth
						{
							position41, tokenIndex41, depth41 := position, tokenIndex, depth
							if buffer[position] != rune('\r') {
								goto l42
							}
							position++
							goto l41
						l42:
							position, tokenIndex, depth = position41, tokenIndex41, depth41
							if buffer[position] != rune('\n') {
								goto l38
							}
							position++
						}
					l41:
						goto l37
					l38:
						position, tokenIndex, depth = position38, tokenIndex38, depth38
					}
					depth--
					add(rulePegText, position36)
				}
				{
					add(ruleAction1, position)
				}
				depth--
				add(rulenewline, position35)
			}
			return true
		l34:
			position, tokenIndex, depth = position34, tokenIndex34, depth34
			return false
		},
		/* 3 ws <- <(' ' / '\t')*> */
		func() bool {
			{
				position45 := position
				depth++
			l46:
				{
					position47, tokenIndex47, depth47 := position, tokenIndex, depth
					{
						position48, tokenIndex48, depth48 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l49
						}
						position++
						goto l48
					l49:
						position, tokenIndex, depth = position48, tokenIndex48, depth48
						if buffer[position] != rune('\t') {
							goto l47
						}
						position++
					}
				l48:
					goto l46
				l47:
					position, tokenIndex, depth = position47, tokenIndex47, depth47
				}
				depth--
				add(rulews, position45)
			}
			return true
		},
		/* 4 wsnl <- <((&('\t') '\t') | (&(' ') ' ') | (&('\n' | '\r') (<('\r' / '\n')> Action2)))*> */
		func() bool {
			{
				position51 := position
				depth++
			l52:
				{
					position53, tokenIndex53, depth53 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '\t':
							if buffer[position] != rune('\t') {
								goto l53
							}
							position++
							break
						case ' ':
							if buffer[position] != rune(' ') {
								goto l53
							}
							position++
							break
						default:
							{
								position55 := position
								depth++
								{
									position56, tokenIndex56, depth56 := position, tokenIndex, depth
									if buffer[position] != rune('\r') {
										goto l57
									}
									position++
									goto l56
								l57:
									position, tokenIndex, depth = position56, tokenIndex56, depth56
									if buffer[position] != rune('\n') {
										goto l53
									}
									position++
								}
							l56:
								depth--
								add(rulePegText, position55)
							}
							{
								add(ruleAction2, position)
							}
							break
						}
					}

					goto l52
				l53:
					position, tokenIndex, depth = position53, tokenIndex53, depth53
				}
				depth--
				add(rulewsnl, position51)
			}
			return true
		},
		/* 5 comment <- <('#' <('\t' / [ -􏿿])*>)> */
		func() bool {
			position59, tokenIndex59, depth59 := position, tokenIndex, depth
			{
				position60 := position
				depth++
				if buffer[position] != rune('#') {
					goto l59
				}
				position++
				{
					position61 := position
					depth++
				l62:
					{
						position63, tokenIndex63, depth63 := position, tokenIndex, depth
						{
							position64, tokenIndex64, depth64 := position, tokenIndex, depth
							if buffer[position] != rune('\t') {
								goto l65
							}
							position++
							goto l64
						l65:
							position, tokenIndex, depth = position64, tokenIndex64, depth64
							if c := buffer[position]; c < rune(' ') || c > rune('\U0010ffff') {
								goto l63
							}
							position++
						}
					l64:
						goto l62
					l63:
						position, tokenIndex, depth = position63, tokenIndex63, depth63
					}
					depth--
					add(rulePegText, position61)
				}
				depth--
				add(rulecomment, position60)
			}
			return true
		l59:
			position, tokenIndex, depth = position59, tokenIndex59, depth59
			return false
		},
		/* 6 keyval <- <(key ws '=' ws val Action3)> */
		func() bool {
			position66, tokenIndex66, depth66 := position, tokenIndex, depth
			{
				position67 := position
				depth++
				if !rules[rulekey]() {
					goto l66
				}
				if !rules[rulews]() {
					goto l66
				}
				if buffer[position] != rune('=') {
					goto l66
				}
				position++
				if !rules[rulews]() {
					goto l66
				}
				if !rules[ruleval]() {
					goto l66
				}
				{
					add(ruleAction3, position)
				}
				depth--
				add(rulekeyval, position67)
			}
			return true
		l66:
			position, tokenIndex, depth = position66, tokenIndex66, depth66
			return false
		},
		/* 7 key <- <(<((&('_') '_') | (&('-') '-') | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]))+> Action4)> */
		func() bool {
			position69, tokenIndex69, depth69 := position, tokenIndex, depth
			{
				position70 := position
				depth++
				{
					position71 := position
					depth++
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l69
							}
							position++
							break
						case '-':
							if buffer[position] != rune('-') {
								goto l69
							}
							position++
							break
						case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l69
							}
							position++
							break
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l69
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l69
							}
							position++
							break
						}
					}

				l72:
					{
						position73, tokenIndex73, depth73 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '_':
								if buffer[position] != rune('_') {
									goto l73
								}
								position++
								break
							case '-':
								if buffer[position] != rune('-') {
									goto l73
								}
								position++
								break
							case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l73
								}
								position++
								break
							case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l73
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l73
								}
								position++
								break
							}
						}

						goto l72
					l73:
						position, tokenIndex, depth = position73, tokenIndex73, depth73
					}
					depth--
					add(rulePegText, position71)
				}
				{
					add(ruleAction4, position)
				}
				depth--
				add(rulekey, position70)
			}
			return true
		l69:
			position, tokenIndex, depth = position69, tokenIndex69, depth69
			return false
		},
		/* 8 val <- <((<datetime> Action5) / (<float> Action6) / ((&('[') (<array> Action10)) | (&('f' | 't') (<boolean> Action9)) | (&('"' | '\'') (<string> Action8)) | (&('-' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') (<integer> Action7))))> */
		func() bool {
			position77, tokenIndex77, depth77 := position, tokenIndex, depth
			{
				position78 := position
				depth++
				{
					position79, tokenIndex79, depth79 := position, tokenIndex, depth
					{
						position81 := position
						depth++
						{
							position82 := position
							depth++
							{
								position83 := position
								depth++
								if !rules[ruledigitDual]() {
									goto l80
								}
								if !rules[ruledigitDual]() {
									goto l80
								}
								depth--
								add(ruledigitQuad, position83)
							}
							if buffer[position] != rune('-') {
								goto l80
							}
							position++
							if !rules[ruledigitDual]() {
								goto l80
							}
							if buffer[position] != rune('-') {
								goto l80
							}
							position++
							if !rules[ruledigitDual]() {
								goto l80
							}
							if buffer[position] != rune('T') {
								goto l80
							}
							position++
							if !rules[ruledigitDual]() {
								goto l80
							}
							if buffer[position] != rune(':') {
								goto l80
							}
							position++
							if !rules[ruledigitDual]() {
								goto l80
							}
							if buffer[position] != rune(':') {
								goto l80
							}
							position++
							if !rules[ruledigitDual]() {
								goto l80
							}
							if buffer[position] != rune('Z') {
								goto l80
							}
							position++
							depth--
							add(ruledatetime, position82)
						}
						depth--
						add(rulePegText, position81)
					}
					{
						add(ruleAction5, position)
					}
					goto l79
				l80:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					{
						position86 := position
						depth++
						{
							position87 := position
							depth++
							{
								position88, tokenIndex88, depth88 := position, tokenIndex, depth
								if buffer[position] != rune('-') {
									goto l88
								}
								position++
								goto l89
							l88:
								position, tokenIndex, depth = position88, tokenIndex88, depth88
							}
						l89:
							if !rules[ruleint]() {
								goto l85
							}
							{
								position90 := position
								depth++
								if buffer[position] != rune('.') {
									goto l85
								}
								position++
								if !rules[ruledigit]() {
									goto l85
								}
							l91:
								{
									position92, tokenIndex92, depth92 := position, tokenIndex, depth
									if !rules[ruledigit]() {
										goto l92
									}
									goto l91
								l92:
									position, tokenIndex, depth = position92, tokenIndex92, depth92
								}
								depth--
								add(rulefrac, position90)
							}
							{
								position93, tokenIndex93, depth93 := position, tokenIndex, depth
								{
									position95 := position
									depth++
									{
										position96, tokenIndex96, depth96 := position, tokenIndex, depth
										if buffer[position] != rune('e') {
											goto l97
										}
										position++
										goto l96
									l97:
										position, tokenIndex, depth = position96, tokenIndex96, depth96
										if buffer[position] != rune('E') {
											goto l93
										}
										position++
									}
								l96:
									{
										position98, tokenIndex98, depth98 := position, tokenIndex, depth
										{
											position100, tokenIndex100, depth100 := position, tokenIndex, depth
											if buffer[position] != rune('-') {
												goto l101
											}
											position++
											goto l100
										l101:
											position, tokenIndex, depth = position100, tokenIndex100, depth100
											if buffer[position] != rune('+') {
												goto l98
											}
											position++
										}
									l100:
										goto l99
									l98:
										position, tokenIndex, depth = position98, tokenIndex98, depth98
									}
								l99:
									if !rules[ruledigit]() {
										goto l93
									}
								l102:
									{
										position103, tokenIndex103, depth103 := position, tokenIndex, depth
										if !rules[ruledigit]() {
											goto l103
										}
										goto l102
									l103:
										position, tokenIndex, depth = position103, tokenIndex103, depth103
									}
									depth--
									add(ruleexp, position95)
								}
								goto l94
							l93:
								position, tokenIndex, depth = position93, tokenIndex93, depth93
							}
						l94:
							depth--
							add(rulefloat, position87)
						}
						depth--
						add(rulePegText, position86)
					}
					{
						add(ruleAction6, position)
					}
					goto l79
				l85:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					{
						switch buffer[position] {
						case '[':
							{
								position106 := position
								depth++
								{
									position107 := position
									depth++
									if buffer[position] != rune('[') {
										goto l77
									}
									position++
									{
										add(ruleAction18, position)
									}
									if !rules[rulewsnl]() {
										goto l77
									}
									{
										position109 := position
										depth++
									l110:
										{
											position111, tokenIndex111, depth111 := position, tokenIndex, depth
											if !rules[ruleval]() {
												goto l111
											}
											{
												add(ruleAction19, position)
											}
											{
												position113, tokenIndex113, depth113 := position, tokenIndex, depth
												{
													position115 := position
													depth++
													if !rules[rulews]() {
														goto l113
													}
													if buffer[position] != rune(',') {
														goto l113
													}
													position++
													if !rules[rulewsnl]() {
														goto l113
													}
													depth--
													add(rulearraySep, position115)
												}
												goto l114
											l113:
												position, tokenIndex, depth = position113, tokenIndex113, depth113
											}
										l114:
											{
												position116, tokenIndex116, depth116 := position, tokenIndex, depth
												{
													position118, tokenIndex118, depth118 := position, tokenIndex, depth
													if !rules[rulecomment]() {
														goto l118
													}
													goto l119
												l118:
													position, tokenIndex, depth = position118, tokenIndex118, depth118
												}
											l119:
												if !rules[rulenewline]() {
													goto l116
												}
												goto l117
											l116:
												position, tokenIndex, depth = position116, tokenIndex116, depth116
											}
										l117:
											goto l110
										l111:
											position, tokenIndex, depth = position111, tokenIndex111, depth111
										}
										depth--
										add(rulearrayValues, position109)
									}
									if !rules[rulewsnl]() {
										goto l77
									}
									if buffer[position] != rune(']') {
										goto l77
									}
									position++
									depth--
									add(rulearray, position107)
								}
								depth--
								add(rulePegText, position106)
							}
							{
								add(ruleAction10, position)
							}
							break
						case 'f', 't':
							{
								position121 := position
								depth++
								{
									position122 := position
									depth++
									{
										position123, tokenIndex123, depth123 := position, tokenIndex, depth
										if buffer[position] != rune('t') {
											goto l124
										}
										position++
										if buffer[position] != rune('r') {
											goto l124
										}
										position++
										if buffer[position] != rune('u') {
											goto l124
										}
										position++
										if buffer[position] != rune('e') {
											goto l124
										}
										position++
										goto l123
									l124:
										position, tokenIndex, depth = position123, tokenIndex123, depth123
										if buffer[position] != rune('f') {
											goto l77
										}
										position++
										if buffer[position] != rune('a') {
											goto l77
										}
										position++
										if buffer[position] != rune('l') {
											goto l77
										}
										position++
										if buffer[position] != rune('s') {
											goto l77
										}
										position++
										if buffer[position] != rune('e') {
											goto l77
										}
										position++
									}
								l123:
									depth--
									add(ruleboolean, position122)
								}
								depth--
								add(rulePegText, position121)
							}
							{
								add(ruleAction9, position)
							}
							break
						case '"', '\'':
							{
								position126 := position
								depth++
								{
									position127 := position
									depth++
									{
										position128, tokenIndex128, depth128 := position, tokenIndex, depth
										{
											position130 := position
											depth++
											if buffer[position] != rune('\'') {
												goto l129
											}
											position++
											if buffer[position] != rune('\'') {
												goto l129
											}
											position++
											if buffer[position] != rune('\'') {
												goto l129
											}
											position++
											{
												position131 := position
												depth++
												{
													position132 := position
													depth++
												l133:
													{
														position134, tokenIndex134, depth134 := position, tokenIndex, depth
														{
															position135, tokenIndex135, depth135 := position, tokenIndex, depth
															if buffer[position] != rune('\'') {
																goto l135
															}
															position++
															if buffer[position] != rune('\'') {
																goto l135
															}
															position++
															if buffer[position] != rune('\'') {
																goto l135
															}
															position++
															goto l134
														l135:
															position, tokenIndex, depth = position135, tokenIndex135, depth135
														}
														{
															position136, tokenIndex136, depth136 := position, tokenIndex, depth
															{
																position138 := position
																depth++
																{
																	position139, tokenIndex139, depth139 := position, tokenIndex, depth
																	if buffer[position] != rune('\t') {
																		goto l140
																	}
																	position++
																	goto l139
																l140:
																	position, tokenIndex, depth = position139, tokenIndex139, depth139
																	if c := buffer[position]; c < rune(' ') || c > rune('\U0010ffff') {
																		goto l137
																	}
																	position++
																}
															l139:
																depth--
																add(rulemlLiteralChar, position138)
															}
															goto l136
														l137:
															position, tokenIndex, depth = position136, tokenIndex136, depth136
															if !rules[rulenewline]() {
																goto l134
															}
														}
													l136:
														goto l133
													l134:
														position, tokenIndex, depth = position134, tokenIndex134, depth134
													}
													depth--
													add(rulemlLiteralBody, position132)
												}
												depth--
												add(rulePegText, position131)
											}
											if buffer[position] != rune('\'') {
												goto l129
											}
											position++
											if buffer[position] != rune('\'') {
												goto l129
											}
											position++
											if buffer[position] != rune('\'') {
												goto l129
											}
											position++
											{
												add(ruleAction17, position)
											}
											depth--
											add(rulemlLiteralString, position130)
										}
										goto l128
									l129:
										position, tokenIndex, depth = position128, tokenIndex128, depth128
										{
											position143 := position
											depth++
											if buffer[position] != rune('\'') {
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
														position147 := position
														depth++
														{
															switch buffer[position] {
															case '\t':
																if buffer[position] != rune('\t') {
																	goto l146
																}
																position++
																break
															case ' ', '!', '"', '#', '$', '%', '&':
																if c := buffer[position]; c < rune(' ') || c > rune('&') {
																	goto l146
																}
																position++
																break
															default:
																if c := buffer[position]; c < rune('(') || c > rune('\U0010ffff') {
																	goto l146
																}
																position++
																break
															}
														}

														depth--
														add(ruleliteralChar, position147)
													}
													goto l145
												l146:
													position, tokenIndex, depth = position146, tokenIndex146, depth146
												}
												depth--
												add(rulePegText, position144)
											}
											if buffer[position] != rune('\'') {
												goto l142
											}
											position++
											{
												add(ruleAction16, position)
											}
											depth--
											add(ruleliteralString, position143)
										}
										goto l128
									l142:
										position, tokenIndex, depth = position128, tokenIndex128, depth128
										{
											position151 := position
											depth++
											if buffer[position] != rune('"') {
												goto l150
											}
											position++
											if buffer[position] != rune('"') {
												goto l150
											}
											position++
											if buffer[position] != rune('"') {
												goto l150
											}
											position++
											{
												position152 := position
												depth++
											l153:
												{
													position154, tokenIndex154, depth154 := position, tokenIndex, depth
													{
														position155, tokenIndex155, depth155 := position, tokenIndex, depth
														{
															position157 := position
															depth++
															{
																position158, tokenIndex158, depth158 := position, tokenIndex, depth
																if !rules[rulebasicChar]() {
																	goto l159
																}
																goto l158
															l159:
																position, tokenIndex, depth = position158, tokenIndex158, depth158
																if !rules[rulenewline]() {
																	goto l156
																}
															}
														l158:
															depth--
															add(rulePegText, position157)
														}
														{
															add(ruleAction15, position)
														}
														goto l155
													l156:
														position, tokenIndex, depth = position155, tokenIndex155, depth155
														if !rules[ruleescape]() {
															goto l154
														}
														if !rules[rulenewline]() {
															goto l154
														}
														if !rules[rulewsnl]() {
															goto l154
														}
													}
												l155:
													goto l153
												l154:
													position, tokenIndex, depth = position154, tokenIndex154, depth154
												}
												depth--
												add(rulemlBasicBody, position152)
											}
											if buffer[position] != rune('"') {
												goto l150
											}
											position++
											if buffer[position] != rune('"') {
												goto l150
											}
											position++
											if buffer[position] != rune('"') {
												goto l150
											}
											position++
											{
												add(ruleAction14, position)
											}
											depth--
											add(rulemlBasicString, position151)
										}
										goto l128
									l150:
										position, tokenIndex, depth = position128, tokenIndex128, depth128
										{
											position162 := position
											depth++
											{
												position163 := position
												depth++
												if buffer[position] != rune('"') {
													goto l77
												}
												position++
											l164:
												{
													position165, tokenIndex165, depth165 := position, tokenIndex, depth
													if !rules[rulebasicChar]() {
														goto l165
													}
													goto l164
												l165:
													position, tokenIndex, depth = position165, tokenIndex165, depth165
												}
												if buffer[position] != rune('"') {
													goto l77
												}
												position++
												depth--
												add(rulePegText, position163)
											}
											{
												add(ruleAction13, position)
											}
											depth--
											add(rulebasicString, position162)
										}
									}
								l128:
									depth--
									add(rulestring, position127)
								}
								depth--
								add(rulePegText, position126)
							}
							{
								add(ruleAction8, position)
							}
							break
						default:
							{
								position168 := position
								depth++
								{
									position169 := position
									depth++
									{
										position170, tokenIndex170, depth170 := position, tokenIndex, depth
										if buffer[position] != rune('-') {
											goto l170
										}
										position++
										goto l171
									l170:
										position, tokenIndex, depth = position170, tokenIndex170, depth170
									}
								l171:
									if !rules[ruleint]() {
										goto l77
									}
									depth--
									add(ruleinteger, position169)
								}
								depth--
								add(rulePegText, position168)
							}
							{
								add(ruleAction7, position)
							}
							break
						}
					}

				}
			l79:
				depth--
				add(ruleval, position78)
			}
			return true
		l77:
			position, tokenIndex, depth = position77, tokenIndex77, depth77
			return false
		},
		/* 9 table <- <(stdTable / arrayTable)> */
		nil,
		/* 10 stdTable <- <('[' ws <tableKey> ws ']' Action11)> */
		nil,
		/* 11 arrayTable <- <('[' '[' ws <tableKey> ws (']' ']') Action12)> */
		nil,
		/* 12 tableKey <- <(key (tableKeySep key)*)> */
		func() bool {
			position176, tokenIndex176, depth176 := position, tokenIndex, depth
			{
				position177 := position
				depth++
				if !rules[rulekey]() {
					goto l176
				}
			l178:
				{
					position179, tokenIndex179, depth179 := position, tokenIndex, depth
					{
						position180 := position
						depth++
						if !rules[rulews]() {
							goto l179
						}
						if buffer[position] != rune('.') {
							goto l179
						}
						position++
						if !rules[rulews]() {
							goto l179
						}
						depth--
						add(ruletableKeySep, position180)
					}
					if !rules[rulekey]() {
						goto l179
					}
					goto l178
				l179:
					position, tokenIndex, depth = position179, tokenIndex179, depth179
				}
				depth--
				add(ruletableKey, position177)
			}
			return true
		l176:
			position, tokenIndex, depth = position176, tokenIndex176, depth176
			return false
		},
		/* 13 tableKeySep <- <(ws '.' ws)> */
		nil,
		/* 14 integer <- <('-'? int)> */
		nil,
		/* 15 int <- <('0' / ([1-9] digit*))> */
		func() bool {
			position183, tokenIndex183, depth183 := position, tokenIndex, depth
			{
				position184 := position
				depth++
				{
					position185, tokenIndex185, depth185 := position, tokenIndex, depth
					if buffer[position] != rune('0') {
						goto l186
					}
					position++
					goto l185
				l186:
					position, tokenIndex, depth = position185, tokenIndex185, depth185
					if c := buffer[position]; c < rune('1') || c > rune('9') {
						goto l183
					}
					position++
				l187:
					{
						position188, tokenIndex188, depth188 := position, tokenIndex, depth
						if !rules[ruledigit]() {
							goto l188
						}
						goto l187
					l188:
						position, tokenIndex, depth = position188, tokenIndex188, depth188
					}
				}
			l185:
				depth--
				add(ruleint, position184)
			}
			return true
		l183:
			position, tokenIndex, depth = position183, tokenIndex183, depth183
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
		/* 20 basicString <- <(<('"' basicChar* '"')> Action13)> */
		nil,
		/* 21 basicChar <- <(basicUnescaped / escaped)> */
		func() bool {
			position194, tokenIndex194, depth194 := position, tokenIndex, depth
			{
				position195 := position
				depth++
				{
					position196, tokenIndex196, depth196 := position, tokenIndex, depth
					{
						position198 := position
						depth++
						{
							switch buffer[position] {
							case ' ', '!':
								if c := buffer[position]; c < rune(' ') || c > rune('!') {
									goto l197
								}
								position++
								break
							case '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ':', ';', '<', '=', '>', '?', '@', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '[':
								if c := buffer[position]; c < rune('#') || c > rune('[') {
									goto l197
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune(']') || c > rune('\U0010ffff') {
									goto l197
								}
								position++
								break
							}
						}

						depth--
						add(rulebasicUnescaped, position198)
					}
					goto l196
				l197:
					position, tokenIndex, depth = position196, tokenIndex196, depth196
					{
						position200 := position
						depth++
						if !rules[ruleescape]() {
							goto l194
						}
						{
							switch buffer[position] {
							case 'U':
								if buffer[position] != rune('U') {
									goto l194
								}
								position++
								if !rules[rulehexQuad]() {
									goto l194
								}
								if !rules[rulehexQuad]() {
									goto l194
								}
								break
							case 'u':
								if buffer[position] != rune('u') {
									goto l194
								}
								position++
								if !rules[rulehexQuad]() {
									goto l194
								}
								break
							case '\\':
								if buffer[position] != rune('\\') {
									goto l194
								}
								position++
								break
							case '/':
								if buffer[position] != rune('/') {
									goto l194
								}
								position++
								break
							case '"':
								if buffer[position] != rune('"') {
									goto l194
								}
								position++
								break
							case 'r':
								if buffer[position] != rune('r') {
									goto l194
								}
								position++
								break
							case 'f':
								if buffer[position] != rune('f') {
									goto l194
								}
								position++
								break
							case 'n':
								if buffer[position] != rune('n') {
									goto l194
								}
								position++
								break
							case 't':
								if buffer[position] != rune('t') {
									goto l194
								}
								position++
								break
							default:
								if buffer[position] != rune('b') {
									goto l194
								}
								position++
								break
							}
						}

						depth--
						add(ruleescaped, position200)
					}
				}
			l196:
				depth--
				add(rulebasicChar, position195)
			}
			return true
		l194:
			position, tokenIndex, depth = position194, tokenIndex194, depth194
			return false
		},
		/* 22 escaped <- <(escape ((&('U') ('U' hexQuad hexQuad)) | (&('u') ('u' hexQuad)) | (&('\\') '\\') | (&('/') '/') | (&('"') '"') | (&('r') 'r') | (&('f') 'f') | (&('n') 'n') | (&('t') 't') | (&('b') 'b')))> */
		nil,
		/* 23 basicUnescaped <- <((&(' ' | '!') [ -!]) | (&('#' | '$' | '%' | '&' | '\'' | '(' | ')' | '*' | '+' | ',' | '-' | '.' | '/' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | ':' | ';' | '<' | '=' | '>' | '?' | '@' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '[') [#-[]) | (&(']' | '^' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z' | '{' | '|' | '}' | '~' | '\u007f' | '\u0080' | '\u0081' | '\u0082' | '\u0083' | '\u0084' | '\u0085' | '\u0086' | '\u0087' | '\u0088' | '\u0089' | '\u008a' | '\u008b' | '\u008c' | '\u008d' | '\u008e' | '\u008f' | '\u0090' | '\u0091' | '\u0092' | '\u0093' | '\u0094' | '\u0095' | '\u0096' | '\u0097' | '\u0098' | '\u0099' | '\u009a' | '\u009b' | '\u009c' | '\u009d' | '\u009e' | '\u009f' | '\u00a0' | '¡' | '¢' | '£' | '¤' | '¥' | '¦' | '§' | '¨' | '©' | 'ª' | '«' | '¬' | '\u00ad' | '®' | '¯' | '°' | '±' | '²' | '³' | '´' | 'µ' | '¶' | '·' | '¸' | '¹' | 'º' | '»' | '¼' | '½' | '¾' | '¿' | 'À' | 'Á' | 'Â' | 'Ã' | 'Ä' | 'Å' | 'Æ' | 'Ç' | 'È' | 'É' | 'Ê' | 'Ë' | 'Ì' | 'Í' | 'Î' | 'Ï' | 'Ð' | 'Ñ' | 'Ò' | 'Ó' | 'Ô' | 'Õ' | 'Ö' | '×' | 'Ø' | 'Ù' | 'Ú' | 'Û' | 'Ü' | 'Ý' | 'Þ' | 'ß' | 'à' | 'á' | 'â' | 'ã' | 'ä' | 'å' | 'æ' | 'ç' | 'è' | 'é' | 'ê' | 'ë' | 'ì' | 'í' | 'î' | 'ï' | 'ð' | 'ñ' | 'ò' | 'ó' | 'ô') []-􏿿]))> */
		nil,
		/* 24 escape <- <'\\'> */
		func() bool {
			position204, tokenIndex204, depth204 := position, tokenIndex, depth
			{
				position205 := position
				depth++
				if buffer[position] != rune('\\') {
					goto l204
				}
				position++
				depth--
				add(ruleescape, position205)
			}
			return true
		l204:
			position, tokenIndex, depth = position204, tokenIndex204, depth204
			return false
		},
		/* 25 mlBasicString <- <('"' '"' '"' mlBasicBody ('"' '"' '"') Action14)> */
		nil,
		/* 26 mlBasicBody <- <((<(basicChar / newline)> Action15) / (escape newline wsnl))*> */
		nil,
		/* 27 literalString <- <('\'' <literalChar*> '\'' Action16)> */
		nil,
		/* 28 literalChar <- <((&('\t') '\t') | (&(' ' | '!' | '"' | '#' | '$' | '%' | '&') [ -&]) | (&('(' | ')' | '*' | '+' | ',' | '-' | '.' | '/' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | ':' | ';' | '<' | '=' | '>' | '?' | '@' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '[' | '\\' | ']' | '^' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z' | '{' | '|' | '}' | '~' | '\u007f' | '\u0080' | '\u0081' | '\u0082' | '\u0083' | '\u0084' | '\u0085' | '\u0086' | '\u0087' | '\u0088' | '\u0089' | '\u008a' | '\u008b' | '\u008c' | '\u008d' | '\u008e' | '\u008f' | '\u0090' | '\u0091' | '\u0092' | '\u0093' | '\u0094' | '\u0095' | '\u0096' | '\u0097' | '\u0098' | '\u0099' | '\u009a' | '\u009b' | '\u009c' | '\u009d' | '\u009e' | '\u009f' | '\u00a0' | '¡' | '¢' | '£' | '¤' | '¥' | '¦' | '§' | '¨' | '©' | 'ª' | '«' | '¬' | '\u00ad' | '®' | '¯' | '°' | '±' | '²' | '³' | '´' | 'µ' | '¶' | '·' | '¸' | '¹' | 'º' | '»' | '¼' | '½' | '¾' | '¿' | 'À' | 'Á' | 'Â' | 'Ã' | 'Ä' | 'Å' | 'Æ' | 'Ç' | 'È' | 'É' | 'Ê' | 'Ë' | 'Ì' | 'Í' | 'Î' | 'Ï' | 'Ð' | 'Ñ' | 'Ò' | 'Ó' | 'Ô' | 'Õ' | 'Ö' | '×' | 'Ø' | 'Ù' | 'Ú' | 'Û' | 'Ü' | 'Ý' | 'Þ' | 'ß' | 'à' | 'á' | 'â' | 'ã' | 'ä' | 'å' | 'æ' | 'ç' | 'è' | 'é' | 'ê' | 'ë' | 'ì' | 'í' | 'î' | 'ï' | 'ð' | 'ñ' | 'ò' | 'ó' | 'ô') [(-􏿿]))> */
		nil,
		/* 29 mlLiteralString <- <('\'' '\'' '\'' <mlLiteralBody> ('\'' '\'' '\'') Action17)> */
		nil,
		/* 30 mlLiteralBody <- <(!('\'' '\'' '\'') (mlLiteralChar / newline))*> */
		nil,
		/* 31 mlLiteralChar <- <('\t' / [ -􏿿])> */
		nil,
		/* 32 hexdigit <- <((&('a' | 'b' | 'c' | 'd' | 'e' | 'f') [a-f]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F') [A-F]) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]))> */
		func() bool {
			position213, tokenIndex213, depth213 := position, tokenIndex, depth
			{
				position214 := position
				depth++
				{
					switch buffer[position] {
					case 'a', 'b', 'c', 'd', 'e', 'f':
						if c := buffer[position]; c < rune('a') || c > rune('f') {
							goto l213
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F':
						if c := buffer[position]; c < rune('A') || c > rune('F') {
							goto l213
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l213
						}
						position++
						break
					}
				}

				depth--
				add(rulehexdigit, position214)
			}
			return true
		l213:
			position, tokenIndex, depth = position213, tokenIndex213, depth213
			return false
		},
		/* 33 hexQuad <- <(hexdigit hexdigit hexdigit hexdigit)> */
		func() bool {
			position216, tokenIndex216, depth216 := position, tokenIndex, depth
			{
				position217 := position
				depth++
				if !rules[rulehexdigit]() {
					goto l216
				}
				if !rules[rulehexdigit]() {
					goto l216
				}
				if !rules[rulehexdigit]() {
					goto l216
				}
				if !rules[rulehexdigit]() {
					goto l216
				}
				depth--
				add(rulehexQuad, position217)
			}
			return true
		l216:
			position, tokenIndex, depth = position216, tokenIndex216, depth216
			return false
		},
		/* 34 boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		nil,
		/* 35 datetime <- <(digitQuad '-' digitDual '-' digitDual 'T' digitDual ':' digitDual ':' digitDual 'Z')> */
		nil,
		/* 36 digit <- <[0-9]> */
		func() bool {
			position220, tokenIndex220, depth220 := position, tokenIndex, depth
			{
				position221 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l220
				}
				position++
				depth--
				add(ruledigit, position221)
			}
			return true
		l220:
			position, tokenIndex, depth = position220, tokenIndex220, depth220
			return false
		},
		/* 37 digitDual <- <(digit digit)> */
		func() bool {
			position222, tokenIndex222, depth222 := position, tokenIndex, depth
			{
				position223 := position
				depth++
				if !rules[ruledigit]() {
					goto l222
				}
				if !rules[ruledigit]() {
					goto l222
				}
				depth--
				add(ruledigitDual, position223)
			}
			return true
		l222:
			position, tokenIndex, depth = position222, tokenIndex222, depth222
			return false
		},
		/* 38 digitQuad <- <(digitDual digitDual)> */
		nil,
		/* 39 array <- <('[' Action18 wsnl arrayValues wsnl ']')> */
		nil,
		/* 40 arrayValues <- <(val Action19 arraySep? (comment? newline)?)*> */
		nil,
		/* 41 arraySep <- <(ws ',' wsnl)> */
		nil,
		nil,
		/* 44 Action0 <- <{ p.SetTableString(begin, end) }> */
		nil,
		/* 45 Action1 <- <{ p.AddLineCount(end - begin) }> */
		nil,
		/* 46 Action2 <- <{ p.AddLineCount(end - begin) }> */
		nil,
		/* 47 Action3 <- <{ p.AddKeyValue() }> */
		nil,
		/* 48 Action4 <- <{ p.SetKey(buffer[begin:end]) }> */
		nil,
		/* 49 Action5 <- <{ p.SetTime(begin, end) }> */
		nil,
		/* 50 Action6 <- <{ p.SetFloat64(begin, end) }> */
		nil,
		/* 51 Action7 <- <{ p.SetInt64(begin, end) }> */
		nil,
		/* 52 Action8 <- <{ p.SetString(begin, end) }> */
		nil,
		/* 53 Action9 <- <{ p.SetBool(begin, end) }> */
		nil,
		/* 54 Action10 <- <{ p.SetArray(begin, end) }> */
		nil,
		/* 55 Action11 <- <{ p.SetTable(buffer[begin:end]) }> */
		nil,
		/* 56 Action12 <- <{ p.SetArrayTable(buffer[begin:end]) }> */
		nil,
		/* 57 Action13 <- <{ p.SetBasicString(p.RuneSlice(buffer, begin, end)) }> */
		nil,
		/* 58 Action14 <- <{ p.SetMultilineString() }> */
		nil,
		/* 59 Action15 <- <{ p.AddMultilineBasicBody(p.RuneSlice(buffer, begin, end)) }> */
		nil,
		/* 60 Action16 <- <{ p.SetLiteralString(p.RuneSlice(buffer, begin, end)) }> */
		nil,
		/* 61 Action17 <- <{ p.SetMultilineLiteralString(p.RuneSlice(buffer, begin, end)) }> */
		nil,
		/* 62 Action18 <- <{ p.StartArray() }> */
		nil,
		/* 63 Action19 <- <{ p.AddArrayVal() }> */
		nil,
	}
	p.rules = rules
}
