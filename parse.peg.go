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
	ruleAction0
	rulePegText
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
	ruleAction20

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
	"Action0",
	"PegText",
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
	"Action20",

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
	rules  [65]func() bool
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
			_ = buffer
		case ruleAction1:
			p.SetTableString(begin, end)
		case ruleAction2:
			p.AddLineCount(end - begin)
		case ruleAction3:
			p.AddLineCount(end - begin)
		case ruleAction4:
			p.AddKeyValue()
		case ruleAction5:
			p.SetKey(p.buffer, begin, end)
		case ruleAction6:
			p.SetTime(begin, end)
		case ruleAction7:
			p.SetFloat64(begin, end)
		case ruleAction8:
			p.SetInt64(begin, end)
		case ruleAction9:
			p.SetString(begin, end)
		case ruleAction10:
			p.SetBool(begin, end)
		case ruleAction11:
			p.SetArray(begin, end)
		case ruleAction12:
			p.SetTable(p.buffer, begin, end)
		case ruleAction13:
			p.SetArrayTable(p.buffer, begin, end)
		case ruleAction14:
			p.SetBasicString(p.buffer, begin, end)
		case ruleAction15:
			p.SetMultilineString()
		case ruleAction16:
			p.AddMultilineBasicBody(p.buffer, begin, end)
		case ruleAction17:
			p.SetLiteralString(p.buffer, begin, end)
		case ruleAction18:
			p.SetMultilineLiteralString(p.buffer, begin, end)
		case ruleAction19:
			p.StartArray()
		case ruleAction20:
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
		/* 0 TOML <- <(Expression (newline Expression)* newline? !. Action0)> */
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
				{
					add(ruleAction0, position)
				}
				depth--
				add(ruleTOML, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 Expression <- <((<(ws table ws comment? (wsnl keyval ws comment?)*)> Action1) / (ws keyval ws comment?) / (ws comment?) / ws)> */
		func() bool {
			position8, tokenIndex8, depth8 := position, tokenIndex, depth
			{
				position9 := position
				depth++
				{
					position10, tokenIndex10, depth10 := position, tokenIndex, depth
					{
						position12 := position
						depth++
						if !rules[rulews]() {
							goto l11
						}
						{
							position13 := position
							depth++
							{
								position14, tokenIndex14, depth14 := position, tokenIndex, depth
								{
									position16 := position
									depth++
									if buffer[position] != rune('[') {
										goto l15
									}
									position++
									if !rules[rulews]() {
										goto l15
									}
									{
										position17 := position
										depth++
										if !rules[ruletableKey]() {
											goto l15
										}
										depth--
										add(rulePegText, position17)
									}
									if !rules[rulews]() {
										goto l15
									}
									if buffer[position] != rune(']') {
										goto l15
									}
									position++
									{
										add(ruleAction12, position)
									}
									depth--
									add(rulestdTable, position16)
								}
								goto l14
							l15:
								position, tokenIndex, depth = position14, tokenIndex14, depth14
								{
									position19 := position
									depth++
									if buffer[position] != rune('[') {
										goto l11
									}
									position++
									if buffer[position] != rune('[') {
										goto l11
									}
									position++
									if !rules[rulews]() {
										goto l11
									}
									{
										position20 := position
										depth++
										if !rules[ruletableKey]() {
											goto l11
										}
										depth--
										add(rulePegText, position20)
									}
									if !rules[rulews]() {
										goto l11
									}
									if buffer[position] != rune(']') {
										goto l11
									}
									position++
									if buffer[position] != rune(']') {
										goto l11
									}
									position++
									{
										add(ruleAction13, position)
									}
									depth--
									add(rulearrayTable, position19)
								}
							}
						l14:
							depth--
							add(ruletable, position13)
						}
						if !rules[rulews]() {
							goto l11
						}
						{
							position22, tokenIndex22, depth22 := position, tokenIndex, depth
							if !rules[rulecomment]() {
								goto l22
							}
							goto l23
						l22:
							position, tokenIndex, depth = position22, tokenIndex22, depth22
						}
					l23:
					l24:
						{
							position25, tokenIndex25, depth25 := position, tokenIndex, depth
							if !rules[rulewsnl]() {
								goto l25
							}
							if !rules[rulekeyval]() {
								goto l25
							}
							if !rules[rulews]() {
								goto l25
							}
							{
								position26, tokenIndex26, depth26 := position, tokenIndex, depth
								if !rules[rulecomment]() {
									goto l26
								}
								goto l27
							l26:
								position, tokenIndex, depth = position26, tokenIndex26, depth26
							}
						l27:
							goto l24
						l25:
							position, tokenIndex, depth = position25, tokenIndex25, depth25
						}
						depth--
						add(rulePegText, position12)
					}
					{
						add(ruleAction1, position)
					}
					goto l10
				l11:
					position, tokenIndex, depth = position10, tokenIndex10, depth10
					if !rules[rulews]() {
						goto l29
					}
					if !rules[rulekeyval]() {
						goto l29
					}
					if !rules[rulews]() {
						goto l29
					}
					{
						position30, tokenIndex30, depth30 := position, tokenIndex, depth
						if !rules[rulecomment]() {
							goto l30
						}
						goto l31
					l30:
						position, tokenIndex, depth = position30, tokenIndex30, depth30
					}
				l31:
					goto l10
				l29:
					position, tokenIndex, depth = position10, tokenIndex10, depth10
					if !rules[rulews]() {
						goto l32
					}
					{
						position33, tokenIndex33, depth33 := position, tokenIndex, depth
						if !rules[rulecomment]() {
							goto l33
						}
						goto l34
					l33:
						position, tokenIndex, depth = position33, tokenIndex33, depth33
					}
				l34:
					goto l10
				l32:
					position, tokenIndex, depth = position10, tokenIndex10, depth10
					if !rules[rulews]() {
						goto l8
					}
				}
			l10:
				depth--
				add(ruleExpression, position9)
			}
			return true
		l8:
			position, tokenIndex, depth = position8, tokenIndex8, depth8
			return false
		},
		/* 2 newline <- <(<('\r' / '\n')+> Action2)> */
		func() bool {
			position35, tokenIndex35, depth35 := position, tokenIndex, depth
			{
				position36 := position
				depth++
				{
					position37 := position
					depth++
					{
						position40, tokenIndex40, depth40 := position, tokenIndex, depth
						if buffer[position] != rune('\r') {
							goto l41
						}
						position++
						goto l40
					l41:
						position, tokenIndex, depth = position40, tokenIndex40, depth40
						if buffer[position] != rune('\n') {
							goto l35
						}
						position++
					}
				l40:
				l38:
					{
						position39, tokenIndex39, depth39 := position, tokenIndex, depth
						{
							position42, tokenIndex42, depth42 := position, tokenIndex, depth
							if buffer[position] != rune('\r') {
								goto l43
							}
							position++
							goto l42
						l43:
							position, tokenIndex, depth = position42, tokenIndex42, depth42
							if buffer[position] != rune('\n') {
								goto l39
							}
							position++
						}
					l42:
						goto l38
					l39:
						position, tokenIndex, depth = position39, tokenIndex39, depth39
					}
					depth--
					add(rulePegText, position37)
				}
				{
					add(ruleAction2, position)
				}
				depth--
				add(rulenewline, position36)
			}
			return true
		l35:
			position, tokenIndex, depth = position35, tokenIndex35, depth35
			return false
		},
		/* 3 ws <- <(' ' / '\t')*> */
		func() bool {
			{
				position46 := position
				depth++
			l47:
				{
					position48, tokenIndex48, depth48 := position, tokenIndex, depth
					{
						position49, tokenIndex49, depth49 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l50
						}
						position++
						goto l49
					l50:
						position, tokenIndex, depth = position49, tokenIndex49, depth49
						if buffer[position] != rune('\t') {
							goto l48
						}
						position++
					}
				l49:
					goto l47
				l48:
					position, tokenIndex, depth = position48, tokenIndex48, depth48
				}
				depth--
				add(rulews, position46)
			}
			return true
		},
		/* 4 wsnl <- <((&('\t') '\t') | (&(' ') ' ') | (&('\n' | '\r') (<('\r' / '\n')> Action3)))*> */
		func() bool {
			{
				position52 := position
				depth++
			l53:
				{
					position54, tokenIndex54, depth54 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '\t':
							if buffer[position] != rune('\t') {
								goto l54
							}
							position++
							break
						case ' ':
							if buffer[position] != rune(' ') {
								goto l54
							}
							position++
							break
						default:
							{
								position56 := position
								depth++
								{
									position57, tokenIndex57, depth57 := position, tokenIndex, depth
									if buffer[position] != rune('\r') {
										goto l58
									}
									position++
									goto l57
								l58:
									position, tokenIndex, depth = position57, tokenIndex57, depth57
									if buffer[position] != rune('\n') {
										goto l54
									}
									position++
								}
							l57:
								depth--
								add(rulePegText, position56)
							}
							{
								add(ruleAction3, position)
							}
							break
						}
					}

					goto l53
				l54:
					position, tokenIndex, depth = position54, tokenIndex54, depth54
				}
				depth--
				add(rulewsnl, position52)
			}
			return true
		},
		/* 5 comment <- <('#' <('\t' / [ -􏿿])*>)> */
		func() bool {
			position60, tokenIndex60, depth60 := position, tokenIndex, depth
			{
				position61 := position
				depth++
				if buffer[position] != rune('#') {
					goto l60
				}
				position++
				{
					position62 := position
					depth++
				l63:
					{
						position64, tokenIndex64, depth64 := position, tokenIndex, depth
						{
							position65, tokenIndex65, depth65 := position, tokenIndex, depth
							if buffer[position] != rune('\t') {
								goto l66
							}
							position++
							goto l65
						l66:
							position, tokenIndex, depth = position65, tokenIndex65, depth65
							if c := buffer[position]; c < rune(' ') || c > rune('\U0010ffff') {
								goto l64
							}
							position++
						}
					l65:
						goto l63
					l64:
						position, tokenIndex, depth = position64, tokenIndex64, depth64
					}
					depth--
					add(rulePegText, position62)
				}
				depth--
				add(rulecomment, position61)
			}
			return true
		l60:
			position, tokenIndex, depth = position60, tokenIndex60, depth60
			return false
		},
		/* 6 keyval <- <(key ws '=' ws val Action4)> */
		func() bool {
			position67, tokenIndex67, depth67 := position, tokenIndex, depth
			{
				position68 := position
				depth++
				if !rules[rulekey]() {
					goto l67
				}
				if !rules[rulews]() {
					goto l67
				}
				if buffer[position] != rune('=') {
					goto l67
				}
				position++
				if !rules[rulews]() {
					goto l67
				}
				if !rules[ruleval]() {
					goto l67
				}
				{
					add(ruleAction4, position)
				}
				depth--
				add(rulekeyval, position68)
			}
			return true
		l67:
			position, tokenIndex, depth = position67, tokenIndex67, depth67
			return false
		},
		/* 7 key <- <(<((&('_') '_') | (&('-') '-') | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]))+> Action5)> */
		func() bool {
			position70, tokenIndex70, depth70 := position, tokenIndex, depth
			{
				position71 := position
				depth++
				{
					position72 := position
					depth++
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l70
							}
							position++
							break
						case '-':
							if buffer[position] != rune('-') {
								goto l70
							}
							position++
							break
						case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l70
							}
							position++
							break
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l70
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l70
							}
							position++
							break
						}
					}

				l73:
					{
						position74, tokenIndex74, depth74 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '_':
								if buffer[position] != rune('_') {
									goto l74
								}
								position++
								break
							case '-':
								if buffer[position] != rune('-') {
									goto l74
								}
								position++
								break
							case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l74
								}
								position++
								break
							case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l74
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l74
								}
								position++
								break
							}
						}

						goto l73
					l74:
						position, tokenIndex, depth = position74, tokenIndex74, depth74
					}
					depth--
					add(rulePegText, position72)
				}
				{
					add(ruleAction5, position)
				}
				depth--
				add(rulekey, position71)
			}
			return true
		l70:
			position, tokenIndex, depth = position70, tokenIndex70, depth70
			return false
		},
		/* 8 val <- <((<datetime> Action6) / (<float> Action7) / ((&('[') (<array> Action11)) | (&('f' | 't') (<boolean> Action10)) | (&('"' | '\'') (<string> Action9)) | (&('-' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') (<integer> Action8))))> */
		func() bool {
			position78, tokenIndex78, depth78 := position, tokenIndex, depth
			{
				position79 := position
				depth++
				{
					position80, tokenIndex80, depth80 := position, tokenIndex, depth
					{
						position82 := position
						depth++
						{
							position83 := position
							depth++
							{
								position84 := position
								depth++
								if !rules[ruledigitDual]() {
									goto l81
								}
								if !rules[ruledigitDual]() {
									goto l81
								}
								depth--
								add(ruledigitQuad, position84)
							}
							if buffer[position] != rune('-') {
								goto l81
							}
							position++
							if !rules[ruledigitDual]() {
								goto l81
							}
							if buffer[position] != rune('-') {
								goto l81
							}
							position++
							if !rules[ruledigitDual]() {
								goto l81
							}
							if buffer[position] != rune('T') {
								goto l81
							}
							position++
							if !rules[ruledigitDual]() {
								goto l81
							}
							if buffer[position] != rune(':') {
								goto l81
							}
							position++
							if !rules[ruledigitDual]() {
								goto l81
							}
							if buffer[position] != rune(':') {
								goto l81
							}
							position++
							if !rules[ruledigitDual]() {
								goto l81
							}
							if buffer[position] != rune('Z') {
								goto l81
							}
							position++
							depth--
							add(ruledatetime, position83)
						}
						depth--
						add(rulePegText, position82)
					}
					{
						add(ruleAction6, position)
					}
					goto l80
				l81:
					position, tokenIndex, depth = position80, tokenIndex80, depth80
					{
						position87 := position
						depth++
						{
							position88 := position
							depth++
							{
								position89, tokenIndex89, depth89 := position, tokenIndex, depth
								if buffer[position] != rune('-') {
									goto l89
								}
								position++
								goto l90
							l89:
								position, tokenIndex, depth = position89, tokenIndex89, depth89
							}
						l90:
							if !rules[ruleint]() {
								goto l86
							}
							{
								position91 := position
								depth++
								if buffer[position] != rune('.') {
									goto l86
								}
								position++
								if !rules[ruledigit]() {
									goto l86
								}
							l92:
								{
									position93, tokenIndex93, depth93 := position, tokenIndex, depth
									if !rules[ruledigit]() {
										goto l93
									}
									goto l92
								l93:
									position, tokenIndex, depth = position93, tokenIndex93, depth93
								}
								depth--
								add(rulefrac, position91)
							}
							{
								position94, tokenIndex94, depth94 := position, tokenIndex, depth
								{
									position96 := position
									depth++
									{
										position97, tokenIndex97, depth97 := position, tokenIndex, depth
										if buffer[position] != rune('e') {
											goto l98
										}
										position++
										goto l97
									l98:
										position, tokenIndex, depth = position97, tokenIndex97, depth97
										if buffer[position] != rune('E') {
											goto l94
										}
										position++
									}
								l97:
									{
										position99, tokenIndex99, depth99 := position, tokenIndex, depth
										{
											position101, tokenIndex101, depth101 := position, tokenIndex, depth
											if buffer[position] != rune('-') {
												goto l102
											}
											position++
											goto l101
										l102:
											position, tokenIndex, depth = position101, tokenIndex101, depth101
											if buffer[position] != rune('+') {
												goto l99
											}
											position++
										}
									l101:
										goto l100
									l99:
										position, tokenIndex, depth = position99, tokenIndex99, depth99
									}
								l100:
									if !rules[ruledigit]() {
										goto l94
									}
								l103:
									{
										position104, tokenIndex104, depth104 := position, tokenIndex, depth
										if !rules[ruledigit]() {
											goto l104
										}
										goto l103
									l104:
										position, tokenIndex, depth = position104, tokenIndex104, depth104
									}
									depth--
									add(ruleexp, position96)
								}
								goto l95
							l94:
								position, tokenIndex, depth = position94, tokenIndex94, depth94
							}
						l95:
							depth--
							add(rulefloat, position88)
						}
						depth--
						add(rulePegText, position87)
					}
					{
						add(ruleAction7, position)
					}
					goto l80
				l86:
					position, tokenIndex, depth = position80, tokenIndex80, depth80
					{
						switch buffer[position] {
						case '[':
							{
								position107 := position
								depth++
								{
									position108 := position
									depth++
									if buffer[position] != rune('[') {
										goto l78
									}
									position++
									{
										add(ruleAction19, position)
									}
									if !rules[rulewsnl]() {
										goto l78
									}
									{
										position110 := position
										depth++
									l111:
										{
											position112, tokenIndex112, depth112 := position, tokenIndex, depth
											if !rules[ruleval]() {
												goto l112
											}
											{
												add(ruleAction20, position)
											}
											{
												position114, tokenIndex114, depth114 := position, tokenIndex, depth
												{
													position116 := position
													depth++
													if !rules[rulews]() {
														goto l114
													}
													if buffer[position] != rune(',') {
														goto l114
													}
													position++
													if !rules[rulewsnl]() {
														goto l114
													}
													depth--
													add(rulearraySep, position116)
												}
												goto l115
											l114:
												position, tokenIndex, depth = position114, tokenIndex114, depth114
											}
										l115:
											{
												position117, tokenIndex117, depth117 := position, tokenIndex, depth
												{
													position119, tokenIndex119, depth119 := position, tokenIndex, depth
													if !rules[rulecomment]() {
														goto l119
													}
													goto l120
												l119:
													position, tokenIndex, depth = position119, tokenIndex119, depth119
												}
											l120:
												if !rules[rulenewline]() {
													goto l117
												}
												goto l118
											l117:
												position, tokenIndex, depth = position117, tokenIndex117, depth117
											}
										l118:
											goto l111
										l112:
											position, tokenIndex, depth = position112, tokenIndex112, depth112
										}
										depth--
										add(rulearrayValues, position110)
									}
									if !rules[rulewsnl]() {
										goto l78
									}
									if buffer[position] != rune(']') {
										goto l78
									}
									position++
									depth--
									add(rulearray, position108)
								}
								depth--
								add(rulePegText, position107)
							}
							{
								add(ruleAction11, position)
							}
							break
						case 'f', 't':
							{
								position122 := position
								depth++
								{
									position123 := position
									depth++
									{
										position124, tokenIndex124, depth124 := position, tokenIndex, depth
										if buffer[position] != rune('t') {
											goto l125
										}
										position++
										if buffer[position] != rune('r') {
											goto l125
										}
										position++
										if buffer[position] != rune('u') {
											goto l125
										}
										position++
										if buffer[position] != rune('e') {
											goto l125
										}
										position++
										goto l124
									l125:
										position, tokenIndex, depth = position124, tokenIndex124, depth124
										if buffer[position] != rune('f') {
											goto l78
										}
										position++
										if buffer[position] != rune('a') {
											goto l78
										}
										position++
										if buffer[position] != rune('l') {
											goto l78
										}
										position++
										if buffer[position] != rune('s') {
											goto l78
										}
										position++
										if buffer[position] != rune('e') {
											goto l78
										}
										position++
									}
								l124:
									depth--
									add(ruleboolean, position123)
								}
								depth--
								add(rulePegText, position122)
							}
							{
								add(ruleAction10, position)
							}
							break
						case '"', '\'':
							{
								position127 := position
								depth++
								{
									position128 := position
									depth++
									{
										position129, tokenIndex129, depth129 := position, tokenIndex, depth
										{
											position131 := position
											depth++
											if buffer[position] != rune('\'') {
												goto l130
											}
											position++
											if buffer[position] != rune('\'') {
												goto l130
											}
											position++
											if buffer[position] != rune('\'') {
												goto l130
											}
											position++
											{
												position132 := position
												depth++
												{
													position133 := position
													depth++
												l134:
													{
														position135, tokenIndex135, depth135 := position, tokenIndex, depth
														{
															position136, tokenIndex136, depth136 := position, tokenIndex, depth
															if buffer[position] != rune('\'') {
																goto l136
															}
															position++
															if buffer[position] != rune('\'') {
																goto l136
															}
															position++
															if buffer[position] != rune('\'') {
																goto l136
															}
															position++
															goto l135
														l136:
															position, tokenIndex, depth = position136, tokenIndex136, depth136
														}
														{
															position137, tokenIndex137, depth137 := position, tokenIndex, depth
															{
																position139 := position
																depth++
																{
																	position140, tokenIndex140, depth140 := position, tokenIndex, depth
																	if buffer[position] != rune('\t') {
																		goto l141
																	}
																	position++
																	goto l140
																l141:
																	position, tokenIndex, depth = position140, tokenIndex140, depth140
																	if c := buffer[position]; c < rune(' ') || c > rune('\U0010ffff') {
																		goto l138
																	}
																	position++
																}
															l140:
																depth--
																add(rulemlLiteralChar, position139)
															}
															goto l137
														l138:
															position, tokenIndex, depth = position137, tokenIndex137, depth137
															if !rules[rulenewline]() {
																goto l135
															}
														}
													l137:
														goto l134
													l135:
														position, tokenIndex, depth = position135, tokenIndex135, depth135
													}
													depth--
													add(rulemlLiteralBody, position133)
												}
												depth--
												add(rulePegText, position132)
											}
											if buffer[position] != rune('\'') {
												goto l130
											}
											position++
											if buffer[position] != rune('\'') {
												goto l130
											}
											position++
											if buffer[position] != rune('\'') {
												goto l130
											}
											position++
											{
												add(ruleAction18, position)
											}
											depth--
											add(rulemlLiteralString, position131)
										}
										goto l129
									l130:
										position, tokenIndex, depth = position129, tokenIndex129, depth129
										{
											position144 := position
											depth++
											if buffer[position] != rune('\'') {
												goto l143
											}
											position++
											{
												position145 := position
												depth++
											l146:
												{
													position147, tokenIndex147, depth147 := position, tokenIndex, depth
													{
														position148 := position
														depth++
														{
															switch buffer[position] {
															case '\t':
																if buffer[position] != rune('\t') {
																	goto l147
																}
																position++
																break
															case ' ', '!', '"', '#', '$', '%', '&':
																if c := buffer[position]; c < rune(' ') || c > rune('&') {
																	goto l147
																}
																position++
																break
															default:
																if c := buffer[position]; c < rune('(') || c > rune('\U0010ffff') {
																	goto l147
																}
																position++
																break
															}
														}

														depth--
														add(ruleliteralChar, position148)
													}
													goto l146
												l147:
													position, tokenIndex, depth = position147, tokenIndex147, depth147
												}
												depth--
												add(rulePegText, position145)
											}
											if buffer[position] != rune('\'') {
												goto l143
											}
											position++
											{
												add(ruleAction17, position)
											}
											depth--
											add(ruleliteralString, position144)
										}
										goto l129
									l143:
										position, tokenIndex, depth = position129, tokenIndex129, depth129
										{
											position152 := position
											depth++
											if buffer[position] != rune('"') {
												goto l151
											}
											position++
											if buffer[position] != rune('"') {
												goto l151
											}
											position++
											if buffer[position] != rune('"') {
												goto l151
											}
											position++
											{
												position153 := position
												depth++
											l154:
												{
													position155, tokenIndex155, depth155 := position, tokenIndex, depth
													{
														position156, tokenIndex156, depth156 := position, tokenIndex, depth
														{
															position158 := position
															depth++
															{
																position159, tokenIndex159, depth159 := position, tokenIndex, depth
																if !rules[rulebasicChar]() {
																	goto l160
																}
																goto l159
															l160:
																position, tokenIndex, depth = position159, tokenIndex159, depth159
																if !rules[rulenewline]() {
																	goto l157
																}
															}
														l159:
															depth--
															add(rulePegText, position158)
														}
														{
															add(ruleAction16, position)
														}
														goto l156
													l157:
														position, tokenIndex, depth = position156, tokenIndex156, depth156
														if !rules[ruleescape]() {
															goto l155
														}
														if !rules[rulenewline]() {
															goto l155
														}
														if !rules[rulewsnl]() {
															goto l155
														}
													}
												l156:
													goto l154
												l155:
													position, tokenIndex, depth = position155, tokenIndex155, depth155
												}
												depth--
												add(rulemlBasicBody, position153)
											}
											if buffer[position] != rune('"') {
												goto l151
											}
											position++
											if buffer[position] != rune('"') {
												goto l151
											}
											position++
											if buffer[position] != rune('"') {
												goto l151
											}
											position++
											{
												add(ruleAction15, position)
											}
											depth--
											add(rulemlBasicString, position152)
										}
										goto l129
									l151:
										position, tokenIndex, depth = position129, tokenIndex129, depth129
										{
											position163 := position
											depth++
											{
												position164 := position
												depth++
												if buffer[position] != rune('"') {
													goto l78
												}
												position++
											l165:
												{
													position166, tokenIndex166, depth166 := position, tokenIndex, depth
													if !rules[rulebasicChar]() {
														goto l166
													}
													goto l165
												l166:
													position, tokenIndex, depth = position166, tokenIndex166, depth166
												}
												if buffer[position] != rune('"') {
													goto l78
												}
												position++
												depth--
												add(rulePegText, position164)
											}
											{
												add(ruleAction14, position)
											}
											depth--
											add(rulebasicString, position163)
										}
									}
								l129:
									depth--
									add(rulestring, position128)
								}
								depth--
								add(rulePegText, position127)
							}
							{
								add(ruleAction9, position)
							}
							break
						default:
							{
								position169 := position
								depth++
								{
									position170 := position
									depth++
									{
										position171, tokenIndex171, depth171 := position, tokenIndex, depth
										if buffer[position] != rune('-') {
											goto l171
										}
										position++
										goto l172
									l171:
										position, tokenIndex, depth = position171, tokenIndex171, depth171
									}
								l172:
									if !rules[ruleint]() {
										goto l78
									}
									depth--
									add(ruleinteger, position170)
								}
								depth--
								add(rulePegText, position169)
							}
							{
								add(ruleAction8, position)
							}
							break
						}
					}

				}
			l80:
				depth--
				add(ruleval, position79)
			}
			return true
		l78:
			position, tokenIndex, depth = position78, tokenIndex78, depth78
			return false
		},
		/* 9 table <- <(stdTable / arrayTable)> */
		nil,
		/* 10 stdTable <- <('[' ws <tableKey> ws ']' Action12)> */
		nil,
		/* 11 arrayTable <- <('[' '[' ws <tableKey> ws (']' ']') Action13)> */
		nil,
		/* 12 tableKey <- <(key (tableKeySep key)*)> */
		func() bool {
			position177, tokenIndex177, depth177 := position, tokenIndex, depth
			{
				position178 := position
				depth++
				if !rules[rulekey]() {
					goto l177
				}
			l179:
				{
					position180, tokenIndex180, depth180 := position, tokenIndex, depth
					{
						position181 := position
						depth++
						if !rules[rulews]() {
							goto l180
						}
						if buffer[position] != rune('.') {
							goto l180
						}
						position++
						if !rules[rulews]() {
							goto l180
						}
						depth--
						add(ruletableKeySep, position181)
					}
					if !rules[rulekey]() {
						goto l180
					}
					goto l179
				l180:
					position, tokenIndex, depth = position180, tokenIndex180, depth180
				}
				depth--
				add(ruletableKey, position178)
			}
			return true
		l177:
			position, tokenIndex, depth = position177, tokenIndex177, depth177
			return false
		},
		/* 13 tableKeySep <- <(ws '.' ws)> */
		nil,
		/* 14 integer <- <('-'? int)> */
		nil,
		/* 15 int <- <('0' / ([1-9] digit*))> */
		func() bool {
			position184, tokenIndex184, depth184 := position, tokenIndex, depth
			{
				position185 := position
				depth++
				{
					position186, tokenIndex186, depth186 := position, tokenIndex, depth
					if buffer[position] != rune('0') {
						goto l187
					}
					position++
					goto l186
				l187:
					position, tokenIndex, depth = position186, tokenIndex186, depth186
					if c := buffer[position]; c < rune('1') || c > rune('9') {
						goto l184
					}
					position++
				l188:
					{
						position189, tokenIndex189, depth189 := position, tokenIndex, depth
						if !rules[ruledigit]() {
							goto l189
						}
						goto l188
					l189:
						position, tokenIndex, depth = position189, tokenIndex189, depth189
					}
				}
			l186:
				depth--
				add(ruleint, position185)
			}
			return true
		l184:
			position, tokenIndex, depth = position184, tokenIndex184, depth184
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
		/* 20 basicString <- <(<('"' basicChar* '"')> Action14)> */
		nil,
		/* 21 basicChar <- <(basicUnescaped / escaped)> */
		func() bool {
			position195, tokenIndex195, depth195 := position, tokenIndex, depth
			{
				position196 := position
				depth++
				{
					position197, tokenIndex197, depth197 := position, tokenIndex, depth
					{
						position199 := position
						depth++
						{
							switch buffer[position] {
							case ' ', '!':
								if c := buffer[position]; c < rune(' ') || c > rune('!') {
									goto l198
								}
								position++
								break
							case '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ':', ';', '<', '=', '>', '?', '@', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '[':
								if c := buffer[position]; c < rune('#') || c > rune('[') {
									goto l198
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune(']') || c > rune('\U0010ffff') {
									goto l198
								}
								position++
								break
							}
						}

						depth--
						add(rulebasicUnescaped, position199)
					}
					goto l197
				l198:
					position, tokenIndex, depth = position197, tokenIndex197, depth197
					{
						position201 := position
						depth++
						if !rules[ruleescape]() {
							goto l195
						}
						{
							switch buffer[position] {
							case 'U':
								if buffer[position] != rune('U') {
									goto l195
								}
								position++
								if !rules[rulehexQuad]() {
									goto l195
								}
								if !rules[rulehexQuad]() {
									goto l195
								}
								break
							case 'u':
								if buffer[position] != rune('u') {
									goto l195
								}
								position++
								if !rules[rulehexQuad]() {
									goto l195
								}
								break
							case '\\':
								if buffer[position] != rune('\\') {
									goto l195
								}
								position++
								break
							case '/':
								if buffer[position] != rune('/') {
									goto l195
								}
								position++
								break
							case '"':
								if buffer[position] != rune('"') {
									goto l195
								}
								position++
								break
							case 'r':
								if buffer[position] != rune('r') {
									goto l195
								}
								position++
								break
							case 'f':
								if buffer[position] != rune('f') {
									goto l195
								}
								position++
								break
							case 'n':
								if buffer[position] != rune('n') {
									goto l195
								}
								position++
								break
							case 't':
								if buffer[position] != rune('t') {
									goto l195
								}
								position++
								break
							default:
								if buffer[position] != rune('b') {
									goto l195
								}
								position++
								break
							}
						}

						depth--
						add(ruleescaped, position201)
					}
				}
			l197:
				depth--
				add(rulebasicChar, position196)
			}
			return true
		l195:
			position, tokenIndex, depth = position195, tokenIndex195, depth195
			return false
		},
		/* 22 escaped <- <(escape ((&('U') ('U' hexQuad hexQuad)) | (&('u') ('u' hexQuad)) | (&('\\') '\\') | (&('/') '/') | (&('"') '"') | (&('r') 'r') | (&('f') 'f') | (&('n') 'n') | (&('t') 't') | (&('b') 'b')))> */
		nil,
		/* 23 basicUnescaped <- <((&(' ' | '!') [ -!]) | (&('#' | '$' | '%' | '&' | '\'' | '(' | ')' | '*' | '+' | ',' | '-' | '.' | '/' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | ':' | ';' | '<' | '=' | '>' | '?' | '@' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '[') [#-[]) | (&(']' | '^' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z' | '{' | '|' | '}' | '~' | '\u007f' | '\u0080' | '\u0081' | '\u0082' | '\u0083' | '\u0084' | '\u0085' | '\u0086' | '\u0087' | '\u0088' | '\u0089' | '\u008a' | '\u008b' | '\u008c' | '\u008d' | '\u008e' | '\u008f' | '\u0090' | '\u0091' | '\u0092' | '\u0093' | '\u0094' | '\u0095' | '\u0096' | '\u0097' | '\u0098' | '\u0099' | '\u009a' | '\u009b' | '\u009c' | '\u009d' | '\u009e' | '\u009f' | '\u00a0' | '¡' | '¢' | '£' | '¤' | '¥' | '¦' | '§' | '¨' | '©' | 'ª' | '«' | '¬' | '\u00ad' | '®' | '¯' | '°' | '±' | '²' | '³' | '´' | 'µ' | '¶' | '·' | '¸' | '¹' | 'º' | '»' | '¼' | '½' | '¾' | '¿' | 'À' | 'Á' | 'Â' | 'Ã' | 'Ä' | 'Å' | 'Æ' | 'Ç' | 'È' | 'É' | 'Ê' | 'Ë' | 'Ì' | 'Í' | 'Î' | 'Ï' | 'Ð' | 'Ñ' | 'Ò' | 'Ó' | 'Ô' | 'Õ' | 'Ö' | '×' | 'Ø' | 'Ù' | 'Ú' | 'Û' | 'Ü' | 'Ý' | 'Þ' | 'ß' | 'à' | 'á' | 'â' | 'ã' | 'ä' | 'å' | 'æ' | 'ç' | 'è' | 'é' | 'ê' | 'ë' | 'ì' | 'í' | 'î' | 'ï' | 'ð' | 'ñ' | 'ò' | 'ó' | 'ô') []-􏿿]))> */
		nil,
		/* 24 escape <- <'\\'> */
		func() bool {
			position205, tokenIndex205, depth205 := position, tokenIndex, depth
			{
				position206 := position
				depth++
				if buffer[position] != rune('\\') {
					goto l205
				}
				position++
				depth--
				add(ruleescape, position206)
			}
			return true
		l205:
			position, tokenIndex, depth = position205, tokenIndex205, depth205
			return false
		},
		/* 25 mlBasicString <- <('"' '"' '"' mlBasicBody ('"' '"' '"') Action15)> */
		nil,
		/* 26 mlBasicBody <- <((<(basicChar / newline)> Action16) / (escape newline wsnl))*> */
		nil,
		/* 27 literalString <- <('\'' <literalChar*> '\'' Action17)> */
		nil,
		/* 28 literalChar <- <((&('\t') '\t') | (&(' ' | '!' | '"' | '#' | '$' | '%' | '&') [ -&]) | (&('(' | ')' | '*' | '+' | ',' | '-' | '.' | '/' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | ':' | ';' | '<' | '=' | '>' | '?' | '@' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '[' | '\\' | ']' | '^' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z' | '{' | '|' | '}' | '~' | '\u007f' | '\u0080' | '\u0081' | '\u0082' | '\u0083' | '\u0084' | '\u0085' | '\u0086' | '\u0087' | '\u0088' | '\u0089' | '\u008a' | '\u008b' | '\u008c' | '\u008d' | '\u008e' | '\u008f' | '\u0090' | '\u0091' | '\u0092' | '\u0093' | '\u0094' | '\u0095' | '\u0096' | '\u0097' | '\u0098' | '\u0099' | '\u009a' | '\u009b' | '\u009c' | '\u009d' | '\u009e' | '\u009f' | '\u00a0' | '¡' | '¢' | '£' | '¤' | '¥' | '¦' | '§' | '¨' | '©' | 'ª' | '«' | '¬' | '\u00ad' | '®' | '¯' | '°' | '±' | '²' | '³' | '´' | 'µ' | '¶' | '·' | '¸' | '¹' | 'º' | '»' | '¼' | '½' | '¾' | '¿' | 'À' | 'Á' | 'Â' | 'Ã' | 'Ä' | 'Å' | 'Æ' | 'Ç' | 'È' | 'É' | 'Ê' | 'Ë' | 'Ì' | 'Í' | 'Î' | 'Ï' | 'Ð' | 'Ñ' | 'Ò' | 'Ó' | 'Ô' | 'Õ' | 'Ö' | '×' | 'Ø' | 'Ù' | 'Ú' | 'Û' | 'Ü' | 'Ý' | 'Þ' | 'ß' | 'à' | 'á' | 'â' | 'ã' | 'ä' | 'å' | 'æ' | 'ç' | 'è' | 'é' | 'ê' | 'ë' | 'ì' | 'í' | 'î' | 'ï' | 'ð' | 'ñ' | 'ò' | 'ó' | 'ô') [(-􏿿]))> */
		nil,
		/* 29 mlLiteralString <- <('\'' '\'' '\'' <mlLiteralBody> ('\'' '\'' '\'') Action18)> */
		nil,
		/* 30 mlLiteralBody <- <(!('\'' '\'' '\'') (mlLiteralChar / newline))*> */
		nil,
		/* 31 mlLiteralChar <- <('\t' / [ -􏿿])> */
		nil,
		/* 32 hexdigit <- <((&('a' | 'b' | 'c' | 'd' | 'e' | 'f') [a-f]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F') [A-F]) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]))> */
		func() bool {
			position214, tokenIndex214, depth214 := position, tokenIndex, depth
			{
				position215 := position
				depth++
				{
					switch buffer[position] {
					case 'a', 'b', 'c', 'd', 'e', 'f':
						if c := buffer[position]; c < rune('a') || c > rune('f') {
							goto l214
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F':
						if c := buffer[position]; c < rune('A') || c > rune('F') {
							goto l214
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l214
						}
						position++
						break
					}
				}

				depth--
				add(rulehexdigit, position215)
			}
			return true
		l214:
			position, tokenIndex, depth = position214, tokenIndex214, depth214
			return false
		},
		/* 33 hexQuad <- <(hexdigit hexdigit hexdigit hexdigit)> */
		func() bool {
			position217, tokenIndex217, depth217 := position, tokenIndex, depth
			{
				position218 := position
				depth++
				if !rules[rulehexdigit]() {
					goto l217
				}
				if !rules[rulehexdigit]() {
					goto l217
				}
				if !rules[rulehexdigit]() {
					goto l217
				}
				if !rules[rulehexdigit]() {
					goto l217
				}
				depth--
				add(rulehexQuad, position218)
			}
			return true
		l217:
			position, tokenIndex, depth = position217, tokenIndex217, depth217
			return false
		},
		/* 34 boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		nil,
		/* 35 datetime <- <(digitQuad '-' digitDual '-' digitDual 'T' digitDual ':' digitDual ':' digitDual 'Z')> */
		nil,
		/* 36 digit <- <[0-9]> */
		func() bool {
			position221, tokenIndex221, depth221 := position, tokenIndex, depth
			{
				position222 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l221
				}
				position++
				depth--
				add(ruledigit, position222)
			}
			return true
		l221:
			position, tokenIndex, depth = position221, tokenIndex221, depth221
			return false
		},
		/* 37 digitDual <- <(digit digit)> */
		func() bool {
			position223, tokenIndex223, depth223 := position, tokenIndex, depth
			{
				position224 := position
				depth++
				if !rules[ruledigit]() {
					goto l223
				}
				if !rules[ruledigit]() {
					goto l223
				}
				depth--
				add(ruledigitDual, position224)
			}
			return true
		l223:
			position, tokenIndex, depth = position223, tokenIndex223, depth223
			return false
		},
		/* 38 digitQuad <- <(digitDual digitDual)> */
		nil,
		/* 39 array <- <('[' Action19 wsnl arrayValues wsnl ']')> */
		nil,
		/* 40 arrayValues <- <(val Action20 arraySep? (comment? newline)?)*> */
		nil,
		/* 41 arraySep <- <(ws ',' wsnl)> */
		nil,
		/* 43 Action0 <- <{ _ = buffer }> */
		nil,
		nil,
		/* 45 Action1 <- <{ p.SetTableString(begin, end) }> */
		nil,
		/* 46 Action2 <- <{ p.AddLineCount(end - begin) }> */
		nil,
		/* 47 Action3 <- <{ p.AddLineCount(end - begin) }> */
		nil,
		/* 48 Action4 <- <{ p.AddKeyValue() }> */
		nil,
		/* 49 Action5 <- <{ p.SetKey(p.buffer, begin, end) }> */
		nil,
		/* 50 Action6 <- <{ p.SetTime(begin, end) }> */
		nil,
		/* 51 Action7 <- <{ p.SetFloat64(begin, end) }> */
		nil,
		/* 52 Action8 <- <{ p.SetInt64(begin, end) }> */
		nil,
		/* 53 Action9 <- <{ p.SetString(begin, end) }> */
		nil,
		/* 54 Action10 <- <{ p.SetBool(begin, end) }> */
		nil,
		/* 55 Action11 <- <{ p.SetArray(begin, end) }> */
		nil,
		/* 56 Action12 <- <{ p.SetTable(p.buffer, begin, end) }> */
		nil,
		/* 57 Action13 <- <{ p.SetArrayTable(p.buffer, begin, end) }> */
		nil,
		/* 58 Action14 <- <{ p.SetBasicString(p.buffer, begin, end) }> */
		nil,
		/* 59 Action15 <- <{ p.SetMultilineString() }> */
		nil,
		/* 60 Action16 <- <{ p.AddMultilineBasicBody(p.buffer, begin, end) }> */
		nil,
		/* 61 Action17 <- <{ p.SetLiteralString(p.buffer, begin, end) }> */
		nil,
		/* 62 Action18 <- <{ p.SetMultilineLiteralString(p.buffer, begin, end) }> */
		nil,
		/* 63 Action19 <- <{ p.StartArray() }> */
		nil,
		/* 64 Action20 <- <{ p.AddArrayVal() }> */
		nil,
	}
	p.rules = rules
}
