package toml

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

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
	rulebareKey
	rulequotedKey
	ruleval
	ruletable
	rulestdTable
	rulearrayTable
	ruleinlineTable
	ruleinlineTableKeyValues
	ruletableKey
	ruletableKeySep
	ruleinlineTableValSep
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
	ruledateFullYear
	ruledateMonth
	ruledateMDay
	ruletimeHour
	ruletimeMinute
	ruletimeSecond
	ruletimeSecfrac
	ruletimeNumoffset
	ruletimeOffset
	rulepartialTime
	rulefullDate
	rulefullTime
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
	ruleAction21
	ruleAction22
	ruleAction23
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
	"bareKey",
	"quotedKey",
	"val",
	"table",
	"stdTable",
	"arrayTable",
	"inlineTable",
	"inlineTableKeyValues",
	"tableKey",
	"tableKeySep",
	"inlineTableValSep",
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
	"dateFullYear",
	"dateMonth",
	"dateMDay",
	"timeHour",
	"timeMinute",
	"timeSecond",
	"timeSecfrac",
	"timeNumoffset",
	"timeOffset",
	"partialTime",
	"fullDate",
	"fullTime",
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
	"Action21",
	"Action22",
	"Action23",
}

type token32 struct {
	pegRule
	begin, end uint32
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v", rul3s[t.pegRule], t.begin, t.end)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(pretty bool, buffer string) {
	var print func(node *node32, depth int)
	print = func(node *node32, depth int) {
		for node != nil {
			for c := 0; c < depth; c++ {
				fmt.Printf(" ")
			}
			rule := rul3s[node.pegRule]
			quote := strconv.Quote(string(([]rune(buffer)[node.begin:node.end])))
			if !pretty {
				fmt.Printf("%v %v\n", rule, quote)
			} else {
				fmt.Printf("\x1B[34m%v\x1B[m %v\n", rule, quote)
			}
			if node.up != nil {
				print(node.up, depth+1)
			}
			node = node.next
		}
	}
	print(node, 0)
}

func (node *node32) Print(buffer string) {
	node.print(false, buffer)
}

func (node *node32) PrettyPrint(buffer string) {
	node.print(true, buffer)
}

type tokens32 struct {
	tree []token32
}

func (t *tokens32) Trim(length uint32) {
	t.tree = t.tree[:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) AST() *node32 {
	type element struct {
		node *node32
		down *element
	}
	tokens := t.Tokens()
	var stack *element
	for _, token := range tokens {
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
	if stack != nil {
		return stack.node
	}
	return nil
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	t.AST().Print(buffer)
}

func (t *tokens32) PrettyPrintSyntaxTree(buffer string) {
	t.AST().PrettyPrint(buffer)
}

func (t *tokens32) Add(rule pegRule, begin, end, index uint32) {
	if tree := t.tree; int(index) >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	t.tree[index] = token32{
		pegRule: rule,
		begin:   begin,
		end:     end,
	}
}

func (t *tokens32) Tokens() []token32 {
	return t.tree
}

type tomlParser struct {
	toml

	Buffer string
	buffer []rune
	rules  [85]func() bool
	parse  func(rule ...int) error
	reset  func()
	Pretty bool
	tokens32
}

func (p *tomlParser) Parse(rule ...int) error {
	return p.parse(rule...)
}

func (p *tomlParser) Reset() {
	p.reset()
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
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
	p   *tomlParser
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return error
}

func (p *tomlParser) PrintSyntaxTree() {
	if p.Pretty {
		p.tokens32.PrettyPrintSyntaxTree(p.Buffer)
	} else {
		p.tokens32.PrintSyntaxTree(p.Buffer)
	}
}

func (p *tomlParser) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for _, token := range p.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])

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
			p.SetKey(p.buffer, begin-1, end+1)
		case ruleAction7:
			p.SetTime(begin, end)
		case ruleAction8:
			p.SetFloat64(begin, end)
		case ruleAction9:
			p.SetInt64(begin, end)
		case ruleAction10:
			p.SetString(begin, end)
		case ruleAction11:
			p.SetBool(begin, end)
		case ruleAction12:
			p.SetArray(begin, end)
		case ruleAction13:
			p.SetTable(p.buffer, begin, end)
		case ruleAction14:
			p.SetArrayTable(p.buffer, begin, end)
		case ruleAction15:
			p.StartInlineTable()
		case ruleAction16:
			p.EndInlineTable()
		case ruleAction17:
			p.SetBasicString(p.buffer, begin, end)
		case ruleAction18:
			p.SetMultilineString()
		case ruleAction19:
			p.AddMultilineBasicBody(p.buffer, begin, end)
		case ruleAction20:
			p.SetLiteralString(p.buffer, begin, end)
		case ruleAction21:
			p.SetMultilineLiteralString(p.buffer, begin, end)
		case ruleAction22:
			p.StartArray()
		case ruleAction23:
			p.AddArrayVal()

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *tomlParser) Init() {
	var (
		max                  token32
		position, tokenIndex uint32
		buffer               []rune
	)
	p.reset = func() {
		max = token32{}
		position, tokenIndex = 0, 0

		p.buffer = []rune(p.Buffer)
		if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
			p.buffer = append(p.buffer, endSymbol)
		}
		buffer = p.buffer
	}
	p.reset()

	_rules := p.rules
	tree := tokens32{tree: make([]token32, math.MaxInt16)}
	p.parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.Trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	add := func(rule pegRule, begin uint32) {
		tree.Add(rule, begin, position, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
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

	_rules = [...]func() bool{
		nil,
		/* 0 TOML <- <(Expression (newline Expression)* newline? !. Action0)> */
		func() bool {
			position0, tokenIndex0 := position, tokenIndex
			{
				position1 := position
				if !_rules[ruleExpression]() {
					goto l0
				}
			l2:
				{
					position3, tokenIndex3 := position, tokenIndex
					if !_rules[rulenewline]() {
						goto l3
					}
					if !_rules[ruleExpression]() {
						goto l3
					}
					goto l2
				l3:
					position, tokenIndex = position3, tokenIndex3
				}
				{
					position4, tokenIndex4 := position, tokenIndex
					if !_rules[rulenewline]() {
						goto l4
					}
					goto l5
				l4:
					position, tokenIndex = position4, tokenIndex4
				}
			l5:
				{
					position6, tokenIndex6 := position, tokenIndex
					if !matchDot() {
						goto l6
					}
					goto l0
				l6:
					position, tokenIndex = position6, tokenIndex6
				}
				{
					add(ruleAction0, position)
				}
				add(ruleTOML, position1)
			}
			return true
		l0:
			position, tokenIndex = position0, tokenIndex0
			return false
		},
		/* 1 Expression <- <((<(ws table ws comment? (wsnl keyval ws comment?)*)> Action1) / (ws keyval ws comment?) / (ws comment?) / ws)> */
		func() bool {
			position8, tokenIndex8 := position, tokenIndex
			{
				position9 := position
				{
					position10, tokenIndex10 := position, tokenIndex
					{
						position12 := position
						if !_rules[rulews]() {
							goto l11
						}
						{
							position13 := position
							{
								position14, tokenIndex14 := position, tokenIndex
								{
									position16 := position
									if buffer[position] != rune('[') {
										goto l15
									}
									position++
									if !_rules[rulews]() {
										goto l15
									}
									{
										position17 := position
										if !_rules[ruletableKey]() {
											goto l15
										}
										add(rulePegText, position17)
									}
									if !_rules[rulews]() {
										goto l15
									}
									if buffer[position] != rune(']') {
										goto l15
									}
									position++
									{
										add(ruleAction13, position)
									}
									add(rulestdTable, position16)
								}
								goto l14
							l15:
								position, tokenIndex = position14, tokenIndex14
								{
									position19 := position
									if buffer[position] != rune('[') {
										goto l11
									}
									position++
									if buffer[position] != rune('[') {
										goto l11
									}
									position++
									if !_rules[rulews]() {
										goto l11
									}
									{
										position20 := position
										if !_rules[ruletableKey]() {
											goto l11
										}
										add(rulePegText, position20)
									}
									if !_rules[rulews]() {
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
										add(ruleAction14, position)
									}
									add(rulearrayTable, position19)
								}
							}
						l14:
							add(ruletable, position13)
						}
						if !_rules[rulews]() {
							goto l11
						}
						{
							position22, tokenIndex22 := position, tokenIndex
							if !_rules[rulecomment]() {
								goto l22
							}
							goto l23
						l22:
							position, tokenIndex = position22, tokenIndex22
						}
					l23:
					l24:
						{
							position25, tokenIndex25 := position, tokenIndex
							if !_rules[rulewsnl]() {
								goto l25
							}
							if !_rules[rulekeyval]() {
								goto l25
							}
							if !_rules[rulews]() {
								goto l25
							}
							{
								position26, tokenIndex26 := position, tokenIndex
								if !_rules[rulecomment]() {
									goto l26
								}
								goto l27
							l26:
								position, tokenIndex = position26, tokenIndex26
							}
						l27:
							goto l24
						l25:
							position, tokenIndex = position25, tokenIndex25
						}
						add(rulePegText, position12)
					}
					{
						add(ruleAction1, position)
					}
					goto l10
				l11:
					position, tokenIndex = position10, tokenIndex10
					if !_rules[rulews]() {
						goto l29
					}
					if !_rules[rulekeyval]() {
						goto l29
					}
					if !_rules[rulews]() {
						goto l29
					}
					{
						position30, tokenIndex30 := position, tokenIndex
						if !_rules[rulecomment]() {
							goto l30
						}
						goto l31
					l30:
						position, tokenIndex = position30, tokenIndex30
					}
				l31:
					goto l10
				l29:
					position, tokenIndex = position10, tokenIndex10
					if !_rules[rulews]() {
						goto l32
					}
					{
						position33, tokenIndex33 := position, tokenIndex
						if !_rules[rulecomment]() {
							goto l33
						}
						goto l34
					l33:
						position, tokenIndex = position33, tokenIndex33
					}
				l34:
					goto l10
				l32:
					position, tokenIndex = position10, tokenIndex10
					if !_rules[rulews]() {
						goto l8
					}
				}
			l10:
				add(ruleExpression, position9)
			}
			return true
		l8:
			position, tokenIndex = position8, tokenIndex8
			return false
		},
		/* 2 newline <- <(<('\r' / '\n')+> Action2)> */
		func() bool {
			position35, tokenIndex35 := position, tokenIndex
			{
				position36 := position
				{
					position37 := position
					{
						position40, tokenIndex40 := position, tokenIndex
						if buffer[position] != rune('\r') {
							goto l41
						}
						position++
						goto l40
					l41:
						position, tokenIndex = position40, tokenIndex40
						if buffer[position] != rune('\n') {
							goto l35
						}
						position++
					}
				l40:
				l38:
					{
						position39, tokenIndex39 := position, tokenIndex
						{
							position42, tokenIndex42 := position, tokenIndex
							if buffer[position] != rune('\r') {
								goto l43
							}
							position++
							goto l42
						l43:
							position, tokenIndex = position42, tokenIndex42
							if buffer[position] != rune('\n') {
								goto l39
							}
							position++
						}
					l42:
						goto l38
					l39:
						position, tokenIndex = position39, tokenIndex39
					}
					add(rulePegText, position37)
				}
				{
					add(ruleAction2, position)
				}
				add(rulenewline, position36)
			}
			return true
		l35:
			position, tokenIndex = position35, tokenIndex35
			return false
		},
		/* 3 ws <- <(' ' / '\t')*> */
		func() bool {
			{
				position46 := position
			l47:
				{
					position48, tokenIndex48 := position, tokenIndex
					{
						position49, tokenIndex49 := position, tokenIndex
						if buffer[position] != rune(' ') {
							goto l50
						}
						position++
						goto l49
					l50:
						position, tokenIndex = position49, tokenIndex49
						if buffer[position] != rune('\t') {
							goto l48
						}
						position++
					}
				l49:
					goto l47
				l48:
					position, tokenIndex = position48, tokenIndex48
				}
				add(rulews, position46)
			}
			return true
		},
		/* 4 wsnl <- <((&('\t') '\t') | (&(' ') ' ') | (&('\n' | '\r') (<('\r' / '\n')> Action3)))*> */
		func() bool {
			{
				position52 := position
			l53:
				{
					position54, tokenIndex54 := position, tokenIndex
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
								{
									position57, tokenIndex57 := position, tokenIndex
									if buffer[position] != rune('\r') {
										goto l58
									}
									position++
									goto l57
								l58:
									position, tokenIndex = position57, tokenIndex57
									if buffer[position] != rune('\n') {
										goto l54
									}
									position++
								}
							l57:
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
					position, tokenIndex = position54, tokenIndex54
				}
				add(rulewsnl, position52)
			}
			return true
		},
		/* 5 comment <- <('#' <('\t' / [ -\U0010ffff])*>)> */
		func() bool {
			position60, tokenIndex60 := position, tokenIndex
			{
				position61 := position
				if buffer[position] != rune('#') {
					goto l60
				}
				position++
				{
					position62 := position
				l63:
					{
						position64, tokenIndex64 := position, tokenIndex
						{
							position65, tokenIndex65 := position, tokenIndex
							if buffer[position] != rune('\t') {
								goto l66
							}
							position++
							goto l65
						l66:
							position, tokenIndex = position65, tokenIndex65
							if c := buffer[position]; c < rune(' ') || c > rune('\U0010ffff') {
								goto l64
							}
							position++
						}
					l65:
						goto l63
					l64:
						position, tokenIndex = position64, tokenIndex64
					}
					add(rulePegText, position62)
				}
				add(rulecomment, position61)
			}
			return true
		l60:
			position, tokenIndex = position60, tokenIndex60
			return false
		},
		/* 6 keyval <- <(key ws '=' ws val Action4)> */
		func() bool {
			position67, tokenIndex67 := position, tokenIndex
			{
				position68 := position
				if !_rules[rulekey]() {
					goto l67
				}
				if !_rules[rulews]() {
					goto l67
				}
				if buffer[position] != rune('=') {
					goto l67
				}
				position++
				if !_rules[rulews]() {
					goto l67
				}
				if !_rules[ruleval]() {
					goto l67
				}
				{
					add(ruleAction4, position)
				}
				add(rulekeyval, position68)
			}
			return true
		l67:
			position, tokenIndex = position67, tokenIndex67
			return false
		},
		/* 7 key <- <(bareKey / quotedKey)> */
		func() bool {
			position70, tokenIndex70 := position, tokenIndex
			{
				position71 := position
				{
					position72, tokenIndex72 := position, tokenIndex
					{
						position74 := position
						{
							position75 := position
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

						l76:
							{
								position77, tokenIndex77 := position, tokenIndex
								{
									switch buffer[position] {
									case '_':
										if buffer[position] != rune('_') {
											goto l77
										}
										position++
										break
									case '-':
										if buffer[position] != rune('-') {
											goto l77
										}
										position++
										break
									case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
										if c := buffer[position]; c < rune('a') || c > rune('z') {
											goto l77
										}
										position++
										break
									case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
										if c := buffer[position]; c < rune('0') || c > rune('9') {
											goto l77
										}
										position++
										break
									default:
										if c := buffer[position]; c < rune('A') || c > rune('Z') {
											goto l77
										}
										position++
										break
									}
								}

								goto l76
							l77:
								position, tokenIndex = position77, tokenIndex77
							}
							add(rulePegText, position75)
						}
						{
							add(ruleAction5, position)
						}
						add(rulebareKey, position74)
					}
					goto l72
				l73:
					position, tokenIndex = position72, tokenIndex72
					{
						position81 := position
						if buffer[position] != rune('"') {
							goto l70
						}
						position++
						{
							position82 := position
							if !_rules[rulebasicChar]() {
								goto l70
							}
						l83:
							{
								position84, tokenIndex84 := position, tokenIndex
								if !_rules[rulebasicChar]() {
									goto l84
								}
								goto l83
							l84:
								position, tokenIndex = position84, tokenIndex84
							}
							add(rulePegText, position82)
						}
						if buffer[position] != rune('"') {
							goto l70
						}
						position++
						{
							add(ruleAction6, position)
						}
						add(rulequotedKey, position81)
					}
				}
			l72:
				add(rulekey, position71)
			}
			return true
		l70:
			position, tokenIndex = position70, tokenIndex70
			return false
		},
		/* 8 bareKey <- <(<((&('_') '_') | (&('-') '-') | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]))+> Action5)> */
		nil,
		/* 9 quotedKey <- <('"' <basicChar+> '"' Action6)> */
		nil,
		/* 10 val <- <((<datetime> Action7) / (<float> Action8) / ((&('{') inlineTable) | (&('[') (<array> Action12)) | (&('f' | 't') (<boolean> Action11)) | (&('"' | '\'') (<string> Action10)) | (&('+' | '-' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') (<integer> Action9))))> */
		func() bool {
			position88, tokenIndex88 := position, tokenIndex
			{
				position89 := position
				{
					position90, tokenIndex90 := position, tokenIndex
					{
						position92 := position
						{
							position93 := position
							{
								position94 := position
								{
									position95 := position
									{
										position96 := position
										if !_rules[ruledigitDual]() {
											goto l91
										}
										if !_rules[ruledigitDual]() {
											goto l91
										}
										add(ruledigitQuad, position96)
									}
									add(ruledateFullYear, position95)
								}
								if buffer[position] != rune('-') {
									goto l91
								}
								position++
								{
									position97 := position
									if !_rules[ruledigitDual]() {
										goto l91
									}
									add(ruledateMonth, position97)
								}
								if buffer[position] != rune('-') {
									goto l91
								}
								position++
								{
									position98 := position
									if !_rules[ruledigitDual]() {
										goto l91
									}
									add(ruledateMDay, position98)
								}
								add(rulefullDate, position94)
							}
							if buffer[position] != rune('T') {
								goto l91
							}
							position++
							{
								position99 := position
								{
									position100 := position
									if !_rules[ruletimeHour]() {
										goto l91
									}
									if buffer[position] != rune(':') {
										goto l91
									}
									position++
									if !_rules[ruletimeMinute]() {
										goto l91
									}
									if buffer[position] != rune(':') {
										goto l91
									}
									position++
									{
										position101 := position
										if !_rules[ruledigitDual]() {
											goto l91
										}
										add(ruletimeSecond, position101)
									}
									{
										position102, tokenIndex102 := position, tokenIndex
										{
											position104 := position
											if buffer[position] != rune('.') {
												goto l102
											}
											position++
											if !_rules[ruledigit]() {
												goto l102
											}
										l105:
											{
												position106, tokenIndex106 := position, tokenIndex
												if !_rules[ruledigit]() {
													goto l106
												}
												goto l105
											l106:
												position, tokenIndex = position106, tokenIndex106
											}
											add(ruletimeSecfrac, position104)
										}
										goto l103
									l102:
										position, tokenIndex = position102, tokenIndex102
									}
								l103:
									add(rulepartialTime, position100)
								}
								{
									position107 := position
									{
										position108, tokenIndex108 := position, tokenIndex
										if buffer[position] != rune('Z') {
											goto l109
										}
										position++
										goto l108
									l109:
										position, tokenIndex = position108, tokenIndex108
										{
											position110 := position
											{
												position111, tokenIndex111 := position, tokenIndex
												if buffer[position] != rune('-') {
													goto l112
												}
												position++
												goto l111
											l112:
												position, tokenIndex = position111, tokenIndex111
												if buffer[position] != rune('+') {
													goto l91
												}
												position++
											}
										l111:
											if !_rules[ruletimeHour]() {
												goto l91
											}
											if buffer[position] != rune(':') {
												goto l91
											}
											position++
											if !_rules[ruletimeMinute]() {
												goto l91
											}
											add(ruletimeNumoffset, position110)
										}
									}
								l108:
									add(ruletimeOffset, position107)
								}
								add(rulefullTime, position99)
							}
							add(ruledatetime, position93)
						}
						add(rulePegText, position92)
					}
					{
						add(ruleAction7, position)
					}
					goto l90
				l91:
					position, tokenIndex = position90, tokenIndex90
					{
						position115 := position
						{
							position116 := position
							if !_rules[ruleinteger]() {
								goto l114
							}
							{
								position117, tokenIndex117 := position, tokenIndex
								if !_rules[rulefrac]() {
									goto l118
								}
								{
									position119, tokenIndex119 := position, tokenIndex
									if !_rules[ruleexp]() {
										goto l119
									}
									goto l120
								l119:
									position, tokenIndex = position119, tokenIndex119
								}
							l120:
								goto l117
							l118:
								position, tokenIndex = position117, tokenIndex117
								{
									position121, tokenIndex121 := position, tokenIndex
									if !_rules[rulefrac]() {
										goto l121
									}
									goto l122
								l121:
									position, tokenIndex = position121, tokenIndex121
								}
							l122:
								if !_rules[ruleexp]() {
									goto l114
								}
							}
						l117:
							add(rulefloat, position116)
						}
						add(rulePegText, position115)
					}
					{
						add(ruleAction8, position)
					}
					goto l90
				l114:
					position, tokenIndex = position90, tokenIndex90
					{
						switch buffer[position] {
						case '{':
							{
								position125 := position
								if buffer[position] != rune('{') {
									goto l88
								}
								position++
								{
									add(ruleAction15, position)
								}
								if !_rules[rulews]() {
									goto l88
								}
								{
									position127 := position
								l128:
									{
										position129, tokenIndex129 := position, tokenIndex
										if !_rules[rulekeyval]() {
											goto l129
										}
										{
											position130, tokenIndex130 := position, tokenIndex
											{
												position132 := position
												if !_rules[rulews]() {
													goto l130
												}
												if buffer[position] != rune(',') {
													goto l130
												}
												position++
												if !_rules[rulews]() {
													goto l130
												}
												add(ruleinlineTableValSep, position132)
											}
											goto l131
										l130:
											position, tokenIndex = position130, tokenIndex130
										}
									l131:
										goto l128
									l129:
										position, tokenIndex = position129, tokenIndex129
									}
									add(ruleinlineTableKeyValues, position127)
								}
								if !_rules[rulews]() {
									goto l88
								}
								if buffer[position] != rune('}') {
									goto l88
								}
								position++
								{
									add(ruleAction16, position)
								}
								add(ruleinlineTable, position125)
							}
							break
						case '[':
							{
								position134 := position
								{
									position135 := position
									if buffer[position] != rune('[') {
										goto l88
									}
									position++
									{
										add(ruleAction22, position)
									}
									if !_rules[rulewsnl]() {
										goto l88
									}
									{
										position137 := position
									l138:
										{
											position139, tokenIndex139 := position, tokenIndex
											if !_rules[ruleval]() {
												goto l139
											}
											{
												add(ruleAction23, position)
											}
											{
												position141, tokenIndex141 := position, tokenIndex
												{
													position143 := position
													if !_rules[rulews]() {
														goto l141
													}
													if buffer[position] != rune(',') {
														goto l141
													}
													position++
													if !_rules[rulewsnl]() {
														goto l141
													}
													add(rulearraySep, position143)
												}
												goto l142
											l141:
												position, tokenIndex = position141, tokenIndex141
											}
										l142:
											{
												position144, tokenIndex144 := position, tokenIndex
												{
													position146, tokenIndex146 := position, tokenIndex
													if !_rules[rulecomment]() {
														goto l146
													}
													goto l147
												l146:
													position, tokenIndex = position146, tokenIndex146
												}
											l147:
												if !_rules[rulenewline]() {
													goto l144
												}
												goto l145
											l144:
												position, tokenIndex = position144, tokenIndex144
											}
										l145:
											goto l138
										l139:
											position, tokenIndex = position139, tokenIndex139
										}
										add(rulearrayValues, position137)
									}
									if !_rules[rulewsnl]() {
										goto l88
									}
									if buffer[position] != rune(']') {
										goto l88
									}
									position++
									add(rulearray, position135)
								}
								add(rulePegText, position134)
							}
							{
								add(ruleAction12, position)
							}
							break
						case 'f', 't':
							{
								position149 := position
								{
									position150 := position
									{
										position151, tokenIndex151 := position, tokenIndex
										if buffer[position] != rune('t') {
											goto l152
										}
										position++
										if buffer[position] != rune('r') {
											goto l152
										}
										position++
										if buffer[position] != rune('u') {
											goto l152
										}
										position++
										if buffer[position] != rune('e') {
											goto l152
										}
										position++
										goto l151
									l152:
										position, tokenIndex = position151, tokenIndex151
										if buffer[position] != rune('f') {
											goto l88
										}
										position++
										if buffer[position] != rune('a') {
											goto l88
										}
										position++
										if buffer[position] != rune('l') {
											goto l88
										}
										position++
										if buffer[position] != rune('s') {
											goto l88
										}
										position++
										if buffer[position] != rune('e') {
											goto l88
										}
										position++
									}
								l151:
									add(ruleboolean, position150)
								}
								add(rulePegText, position149)
							}
							{
								add(ruleAction11, position)
							}
							break
						case '"', '\'':
							{
								position154 := position
								{
									position155 := position
									{
										position156, tokenIndex156 := position, tokenIndex
										{
											position158 := position
											if buffer[position] != rune('\'') {
												goto l157
											}
											position++
											if buffer[position] != rune('\'') {
												goto l157
											}
											position++
											if buffer[position] != rune('\'') {
												goto l157
											}
											position++
											{
												position159 := position
												{
													position160 := position
												l161:
													{
														position162, tokenIndex162 := position, tokenIndex
														{
															position163, tokenIndex163 := position, tokenIndex
															if buffer[position] != rune('\'') {
																goto l163
															}
															position++
															if buffer[position] != rune('\'') {
																goto l163
															}
															position++
															if buffer[position] != rune('\'') {
																goto l163
															}
															position++
															goto l162
														l163:
															position, tokenIndex = position163, tokenIndex163
														}
														{
															position164, tokenIndex164 := position, tokenIndex
															{
																position166 := position
																{
																	position167, tokenIndex167 := position, tokenIndex
																	if buffer[position] != rune('\t') {
																		goto l168
																	}
																	position++
																	goto l167
																l168:
																	position, tokenIndex = position167, tokenIndex167
																	if c := buffer[position]; c < rune(' ') || c > rune('\U0010ffff') {
																		goto l165
																	}
																	position++
																}
															l167:
																add(rulemlLiteralChar, position166)
															}
															goto l164
														l165:
															position, tokenIndex = position164, tokenIndex164
															if !_rules[rulenewline]() {
																goto l162
															}
														}
													l164:
														goto l161
													l162:
														position, tokenIndex = position162, tokenIndex162
													}
													add(rulemlLiteralBody, position160)
												}
												add(rulePegText, position159)
											}
											if buffer[position] != rune('\'') {
												goto l157
											}
											position++
											if buffer[position] != rune('\'') {
												goto l157
											}
											position++
											if buffer[position] != rune('\'') {
												goto l157
											}
											position++
											{
												add(ruleAction21, position)
											}
											add(rulemlLiteralString, position158)
										}
										goto l156
									l157:
										position, tokenIndex = position156, tokenIndex156
										{
											position171 := position
											if buffer[position] != rune('\'') {
												goto l170
											}
											position++
											{
												position172 := position
											l173:
												{
													position174, tokenIndex174 := position, tokenIndex
													{
														position175 := position
														{
															switch buffer[position] {
															case '\t':
																if buffer[position] != rune('\t') {
																	goto l174
																}
																position++
																break
															case ' ', '!', '"', '#', '$', '%', '&':
																if c := buffer[position]; c < rune(' ') || c > rune('&') {
																	goto l174
																}
																position++
																break
															default:
																if c := buffer[position]; c < rune('(') || c > rune('\U0010ffff') {
																	goto l174
																}
																position++
																break
															}
														}

														add(ruleliteralChar, position175)
													}
													goto l173
												l174:
													position, tokenIndex = position174, tokenIndex174
												}
												add(rulePegText, position172)
											}
											if buffer[position] != rune('\'') {
												goto l170
											}
											position++
											{
												add(ruleAction20, position)
											}
											add(ruleliteralString, position171)
										}
										goto l156
									l170:
										position, tokenIndex = position156, tokenIndex156
										{
											position179 := position
											if buffer[position] != rune('"') {
												goto l178
											}
											position++
											if buffer[position] != rune('"') {
												goto l178
											}
											position++
											if buffer[position] != rune('"') {
												goto l178
											}
											position++
											{
												position180 := position
											l181:
												{
													position182, tokenIndex182 := position, tokenIndex
													{
														position183, tokenIndex183 := position, tokenIndex
														{
															position185 := position
															{
																position186, tokenIndex186 := position, tokenIndex
																if !_rules[rulebasicChar]() {
																	goto l187
																}
																goto l186
															l187:
																position, tokenIndex = position186, tokenIndex186
																if !_rules[rulenewline]() {
																	goto l184
																}
															}
														l186:
															add(rulePegText, position185)
														}
														{
															add(ruleAction19, position)
														}
														goto l183
													l184:
														position, tokenIndex = position183, tokenIndex183
														if !_rules[ruleescape]() {
															goto l182
														}
														if !_rules[rulenewline]() {
															goto l182
														}
														if !_rules[rulewsnl]() {
															goto l182
														}
													}
												l183:
													goto l181
												l182:
													position, tokenIndex = position182, tokenIndex182
												}
												add(rulemlBasicBody, position180)
											}
											if buffer[position] != rune('"') {
												goto l178
											}
											position++
											if buffer[position] != rune('"') {
												goto l178
											}
											position++
											if buffer[position] != rune('"') {
												goto l178
											}
											position++
											{
												add(ruleAction18, position)
											}
											add(rulemlBasicString, position179)
										}
										goto l156
									l178:
										position, tokenIndex = position156, tokenIndex156
										{
											position190 := position
											{
												position191 := position
												if buffer[position] != rune('"') {
													goto l88
												}
												position++
											l192:
												{
													position193, tokenIndex193 := position, tokenIndex
													if !_rules[rulebasicChar]() {
														goto l193
													}
													goto l192
												l193:
													position, tokenIndex = position193, tokenIndex193
												}
												if buffer[position] != rune('"') {
													goto l88
												}
												position++
												add(rulePegText, position191)
											}
											{
												add(ruleAction17, position)
											}
											add(rulebasicString, position190)
										}
									}
								l156:
									add(rulestring, position155)
								}
								add(rulePegText, position154)
							}
							{
								add(ruleAction10, position)
							}
							break
						default:
							{
								position196 := position
								if !_rules[ruleinteger]() {
									goto l88
								}
								add(rulePegText, position196)
							}
							{
								add(ruleAction9, position)
							}
							break
						}
					}

				}
			l90:
				add(ruleval, position89)
			}
			return true
		l88:
			position, tokenIndex = position88, tokenIndex88
			return false
		},
		/* 11 table <- <(stdTable / arrayTable)> */
		nil,
		/* 12 stdTable <- <('[' ws <tableKey> ws ']' Action13)> */
		nil,
		/* 13 arrayTable <- <('[' '[' ws <tableKey> ws (']' ']') Action14)> */
		nil,
		/* 14 inlineTable <- <('{' Action15 ws inlineTableKeyValues ws '}' Action16)> */
		nil,
		/* 15 inlineTableKeyValues <- <(keyval inlineTableValSep?)*> */
		nil,
		/* 16 tableKey <- <(key (tableKeySep key)*)> */
		func() bool {
			position203, tokenIndex203 := position, tokenIndex
			{
				position204 := position
				if !_rules[rulekey]() {
					goto l203
				}
			l205:
				{
					position206, tokenIndex206 := position, tokenIndex
					{
						position207 := position
						if !_rules[rulews]() {
							goto l206
						}
						if buffer[position] != rune('.') {
							goto l206
						}
						position++
						if !_rules[rulews]() {
							goto l206
						}
						add(ruletableKeySep, position207)
					}
					if !_rules[rulekey]() {
						goto l206
					}
					goto l205
				l206:
					position, tokenIndex = position206, tokenIndex206
				}
				add(ruletableKey, position204)
			}
			return true
		l203:
			position, tokenIndex = position203, tokenIndex203
			return false
		},
		/* 17 tableKeySep <- <(ws '.' ws)> */
		nil,
		/* 18 inlineTableValSep <- <(ws ',' ws)> */
		nil,
		/* 19 integer <- <(('-' / '+')? int)> */
		func() bool {
			position210, tokenIndex210 := position, tokenIndex
			{
				position211 := position
				{
					position212, tokenIndex212 := position, tokenIndex
					{
						position214, tokenIndex214 := position, tokenIndex
						if buffer[position] != rune('-') {
							goto l215
						}
						position++
						goto l214
					l215:
						position, tokenIndex = position214, tokenIndex214
						if buffer[position] != rune('+') {
							goto l212
						}
						position++
					}
				l214:
					goto l213
				l212:
					position, tokenIndex = position212, tokenIndex212
				}
			l213:
				{
					position216 := position
					{
						position217, tokenIndex217 := position, tokenIndex
						if c := buffer[position]; c < rune('1') || c > rune('9') {
							goto l218
						}
						position++
						{
							position221, tokenIndex221 := position, tokenIndex
							if !_rules[ruledigit]() {
								goto l222
							}
							goto l221
						l222:
							position, tokenIndex = position221, tokenIndex221
							if buffer[position] != rune('_') {
								goto l218
							}
							position++
							if !_rules[ruledigit]() {
								goto l218
							}
						}
					l221:
					l219:
						{
							position220, tokenIndex220 := position, tokenIndex
							{
								position223, tokenIndex223 := position, tokenIndex
								if !_rules[ruledigit]() {
									goto l224
								}
								goto l223
							l224:
								position, tokenIndex = position223, tokenIndex223
								if buffer[position] != rune('_') {
									goto l220
								}
								position++
								if !_rules[ruledigit]() {
									goto l220
								}
							}
						l223:
							goto l219
						l220:
							position, tokenIndex = position220, tokenIndex220
						}
						goto l217
					l218:
						position, tokenIndex = position217, tokenIndex217
						if !_rules[ruledigit]() {
							goto l210
						}
					}
				l217:
					add(ruleint, position216)
				}
				add(ruleinteger, position211)
			}
			return true
		l210:
			position, tokenIndex = position210, tokenIndex210
			return false
		},
		/* 20 int <- <(([1-9] (digit / ('_' digit))+) / digit)> */
		nil,
		/* 21 float <- <(integer ((frac exp?) / (frac? exp)))> */
		nil,
		/* 22 frac <- <('.' digit (digit / ('_' digit))*)> */
		func() bool {
			position227, tokenIndex227 := position, tokenIndex
			{
				position228 := position
				if buffer[position] != rune('.') {
					goto l227
				}
				position++
				if !_rules[ruledigit]() {
					goto l227
				}
			l229:
				{
					position230, tokenIndex230 := position, tokenIndex
					{
						position231, tokenIndex231 := position, tokenIndex
						if !_rules[ruledigit]() {
							goto l232
						}
						goto l231
					l232:
						position, tokenIndex = position231, tokenIndex231
						if buffer[position] != rune('_') {
							goto l230
						}
						position++
						if !_rules[ruledigit]() {
							goto l230
						}
					}
				l231:
					goto l229
				l230:
					position, tokenIndex = position230, tokenIndex230
				}
				add(rulefrac, position228)
			}
			return true
		l227:
			position, tokenIndex = position227, tokenIndex227
			return false
		},
		/* 23 exp <- <(('e' / 'E') ('-' / '+')? digit (digit / ('_' digit))*)> */
		func() bool {
			position233, tokenIndex233 := position, tokenIndex
			{
				position234 := position
				{
					position235, tokenIndex235 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l236
					}
					position++
					goto l235
				l236:
					position, tokenIndex = position235, tokenIndex235
					if buffer[position] != rune('E') {
						goto l233
					}
					position++
				}
			l235:
				{
					position237, tokenIndex237 := position, tokenIndex
					{
						position239, tokenIndex239 := position, tokenIndex
						if buffer[position] != rune('-') {
							goto l240
						}
						position++
						goto l239
					l240:
						position, tokenIndex = position239, tokenIndex239
						if buffer[position] != rune('+') {
							goto l237
						}
						position++
					}
				l239:
					goto l238
				l237:
					position, tokenIndex = position237, tokenIndex237
				}
			l238:
				if !_rules[ruledigit]() {
					goto l233
				}
			l241:
				{
					position242, tokenIndex242 := position, tokenIndex
					{
						position243, tokenIndex243 := position, tokenIndex
						if !_rules[ruledigit]() {
							goto l244
						}
						goto l243
					l244:
						position, tokenIndex = position243, tokenIndex243
						if buffer[position] != rune('_') {
							goto l242
						}
						position++
						if !_rules[ruledigit]() {
							goto l242
						}
					}
				l243:
					goto l241
				l242:
					position, tokenIndex = position242, tokenIndex242
				}
				add(ruleexp, position234)
			}
			return true
		l233:
			position, tokenIndex = position233, tokenIndex233
			return false
		},
		/* 24 string <- <(mlLiteralString / literalString / mlBasicString / basicString)> */
		nil,
		/* 25 basicString <- <(<('"' basicChar* '"')> Action17)> */
		nil,
		/* 26 basicChar <- <(basicUnescaped / escaped)> */
		func() bool {
			position247, tokenIndex247 := position, tokenIndex
			{
				position248 := position
				{
					position249, tokenIndex249 := position, tokenIndex
					{
						position251 := position
						{
							switch buffer[position] {
							case ' ', '!':
								if c := buffer[position]; c < rune(' ') || c > rune('!') {
									goto l250
								}
								position++
								break
							case '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ':', ';', '<', '=', '>', '?', '@', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '[':
								if c := buffer[position]; c < rune('#') || c > rune('[') {
									goto l250
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune(']') || c > rune('\U0010ffff') {
									goto l250
								}
								position++
								break
							}
						}

						add(rulebasicUnescaped, position251)
					}
					goto l249
				l250:
					position, tokenIndex = position249, tokenIndex249
					{
						position253 := position
						if !_rules[ruleescape]() {
							goto l247
						}
						{
							switch buffer[position] {
							case 'U':
								if buffer[position] != rune('U') {
									goto l247
								}
								position++
								if !_rules[rulehexQuad]() {
									goto l247
								}
								if !_rules[rulehexQuad]() {
									goto l247
								}
								break
							case 'u':
								if buffer[position] != rune('u') {
									goto l247
								}
								position++
								if !_rules[rulehexQuad]() {
									goto l247
								}
								break
							case '\\':
								if buffer[position] != rune('\\') {
									goto l247
								}
								position++
								break
							case '/':
								if buffer[position] != rune('/') {
									goto l247
								}
								position++
								break
							case '"':
								if buffer[position] != rune('"') {
									goto l247
								}
								position++
								break
							case 'r':
								if buffer[position] != rune('r') {
									goto l247
								}
								position++
								break
							case 'f':
								if buffer[position] != rune('f') {
									goto l247
								}
								position++
								break
							case 'n':
								if buffer[position] != rune('n') {
									goto l247
								}
								position++
								break
							case 't':
								if buffer[position] != rune('t') {
									goto l247
								}
								position++
								break
							default:
								if buffer[position] != rune('b') {
									goto l247
								}
								position++
								break
							}
						}

						add(ruleescaped, position253)
					}
				}
			l249:
				add(rulebasicChar, position248)
			}
			return true
		l247:
			position, tokenIndex = position247, tokenIndex247
			return false
		},
		/* 27 escaped <- <(escape ((&('U') ('U' hexQuad hexQuad)) | (&('u') ('u' hexQuad)) | (&('\\') '\\') | (&('/') '/') | (&('"') '"') | (&('r') 'r') | (&('f') 'f') | (&('n') 'n') | (&('t') 't') | (&('b') 'b')))> */
		nil,
		/* 28 basicUnescaped <- <((&(' ' | '!') [ -!]) | (&('#' | '$' | '%' | '&' | '\'' | '(' | ')' | '*' | '+' | ',' | '-' | '.' | '/' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | ':' | ';' | '<' | '=' | '>' | '?' | '@' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '[') [#-[]) | (&(']' | '^' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z' | '{' | '|' | '}' | '~' | '\u007f' | '\u0080' | '\u0081' | '\u0082' | '\u0083' | '\u0084' | '\u0085' | '\u0086' | '\u0087' | '\u0088' | '\u0089' | '\u008a' | '\u008b' | '\u008c' | '\u008d' | '\u008e' | '\u008f' | '\u0090' | '\u0091' | '\u0092' | '\u0093' | '\u0094' | '\u0095' | '\u0096' | '\u0097' | '\u0098' | '\u0099' | '\u009a' | '\u009b' | '\u009c' | '\u009d' | '\u009e' | '\u009f' | '\u00a0' | '¡' | '¢' | '£' | '¤' | '¥' | '¦' | '§' | '¨' | '©' | 'ª' | '«' | '¬' | '\u00ad' | '®' | '¯' | '°' | '±' | '²' | '³' | '´' | 'µ' | '¶' | '·' | '¸' | '¹' | 'º' | '»' | '¼' | '½' | '¾' | '¿' | 'À' | 'Á' | 'Â' | 'Ã' | 'Ä' | 'Å' | 'Æ' | 'Ç' | 'È' | 'É' | 'Ê' | 'Ë' | 'Ì' | 'Í' | 'Î' | 'Ï' | 'Ð' | 'Ñ' | 'Ò' | 'Ó' | 'Ô' | 'Õ' | 'Ö' | '×' | 'Ø' | 'Ù' | 'Ú' | 'Û' | 'Ü' | 'Ý' | 'Þ' | 'ß' | 'à' | 'á' | 'â' | 'ã' | 'ä' | 'å' | 'æ' | 'ç' | 'è' | 'é' | 'ê' | 'ë' | 'ì' | 'í' | 'î' | 'ï' | 'ð' | 'ñ' | 'ò' | 'ó' | 'ô' | 'õ' | 'ö' | '÷' | 'ø' | 'ù' | 'ú' | 'û' | 'ü' | 'ý' | 'þ' | 'ÿ') []-\U0010ffff]))> */
		nil,
		/* 29 escape <- <'\\'> */
		func() bool {
			position257, tokenIndex257 := position, tokenIndex
			{
				position258 := position
				if buffer[position] != rune('\\') {
					goto l257
				}
				position++
				add(ruleescape, position258)
			}
			return true
		l257:
			position, tokenIndex = position257, tokenIndex257
			return false
		},
		/* 30 mlBasicString <- <('"' '"' '"' mlBasicBody ('"' '"' '"') Action18)> */
		nil,
		/* 31 mlBasicBody <- <((<(basicChar / newline)> Action19) / (escape newline wsnl))*> */
		nil,
		/* 32 literalString <- <('\'' <literalChar*> '\'' Action20)> */
		nil,
		/* 33 literalChar <- <((&('\t') '\t') | (&(' ' | '!' | '"' | '#' | '$' | '%' | '&') [ -&]) | (&('(' | ')' | '*' | '+' | ',' | '-' | '.' | '/' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | ':' | ';' | '<' | '=' | '>' | '?' | '@' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '[' | '\\' | ']' | '^' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z' | '{' | '|' | '}' | '~' | '\u007f' | '\u0080' | '\u0081' | '\u0082' | '\u0083' | '\u0084' | '\u0085' | '\u0086' | '\u0087' | '\u0088' | '\u0089' | '\u008a' | '\u008b' | '\u008c' | '\u008d' | '\u008e' | '\u008f' | '\u0090' | '\u0091' | '\u0092' | '\u0093' | '\u0094' | '\u0095' | '\u0096' | '\u0097' | '\u0098' | '\u0099' | '\u009a' | '\u009b' | '\u009c' | '\u009d' | '\u009e' | '\u009f' | '\u00a0' | '¡' | '¢' | '£' | '¤' | '¥' | '¦' | '§' | '¨' | '©' | 'ª' | '«' | '¬' | '\u00ad' | '®' | '¯' | '°' | '±' | '²' | '³' | '´' | 'µ' | '¶' | '·' | '¸' | '¹' | 'º' | '»' | '¼' | '½' | '¾' | '¿' | 'À' | 'Á' | 'Â' | 'Ã' | 'Ä' | 'Å' | 'Æ' | 'Ç' | 'È' | 'É' | 'Ê' | 'Ë' | 'Ì' | 'Í' | 'Î' | 'Ï' | 'Ð' | 'Ñ' | 'Ò' | 'Ó' | 'Ô' | 'Õ' | 'Ö' | '×' | 'Ø' | 'Ù' | 'Ú' | 'Û' | 'Ü' | 'Ý' | 'Þ' | 'ß' | 'à' | 'á' | 'â' | 'ã' | 'ä' | 'å' | 'æ' | 'ç' | 'è' | 'é' | 'ê' | 'ë' | 'ì' | 'í' | 'î' | 'ï' | 'ð' | 'ñ' | 'ò' | 'ó' | 'ô' | 'õ' | 'ö' | '÷' | 'ø' | 'ù' | 'ú' | 'û' | 'ü' | 'ý' | 'þ' | 'ÿ') [(-\U0010ffff]))> */
		nil,
		/* 34 mlLiteralString <- <('\'' '\'' '\'' <mlLiteralBody> ('\'' '\'' '\'') Action21)> */
		nil,
		/* 35 mlLiteralBody <- <(!('\'' '\'' '\'') (mlLiteralChar / newline))*> */
		nil,
		/* 36 mlLiteralChar <- <('\t' / [ -\U0010ffff])> */
		nil,
		/* 37 hexdigit <- <((&('a' | 'b' | 'c' | 'd' | 'e' | 'f') [a-f]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F') [A-F]) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]))> */
		func() bool {
			position266, tokenIndex266 := position, tokenIndex
			{
				position267 := position
				{
					switch buffer[position] {
					case 'a', 'b', 'c', 'd', 'e', 'f':
						if c := buffer[position]; c < rune('a') || c > rune('f') {
							goto l266
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F':
						if c := buffer[position]; c < rune('A') || c > rune('F') {
							goto l266
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l266
						}
						position++
						break
					}
				}

				add(rulehexdigit, position267)
			}
			return true
		l266:
			position, tokenIndex = position266, tokenIndex266
			return false
		},
		/* 38 hexQuad <- <(hexdigit hexdigit hexdigit hexdigit)> */
		func() bool {
			position269, tokenIndex269 := position, tokenIndex
			{
				position270 := position
				if !_rules[rulehexdigit]() {
					goto l269
				}
				if !_rules[rulehexdigit]() {
					goto l269
				}
				if !_rules[rulehexdigit]() {
					goto l269
				}
				if !_rules[rulehexdigit]() {
					goto l269
				}
				add(rulehexQuad, position270)
			}
			return true
		l269:
			position, tokenIndex = position269, tokenIndex269
			return false
		},
		/* 39 boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		nil,
		/* 40 dateFullYear <- <digitQuad> */
		nil,
		/* 41 dateMonth <- <digitDual> */
		nil,
		/* 42 dateMDay <- <digitDual> */
		nil,
		/* 43 timeHour <- <digitDual> */
		func() bool {
			position275, tokenIndex275 := position, tokenIndex
			{
				position276 := position
				if !_rules[ruledigitDual]() {
					goto l275
				}
				add(ruletimeHour, position276)
			}
			return true
		l275:
			position, tokenIndex = position275, tokenIndex275
			return false
		},
		/* 44 timeMinute <- <digitDual> */
		func() bool {
			position277, tokenIndex277 := position, tokenIndex
			{
				position278 := position
				if !_rules[ruledigitDual]() {
					goto l277
				}
				add(ruletimeMinute, position278)
			}
			return true
		l277:
			position, tokenIndex = position277, tokenIndex277
			return false
		},
		/* 45 timeSecond <- <digitDual> */
		nil,
		/* 46 timeSecfrac <- <('.' digit+)> */
		nil,
		/* 47 timeNumoffset <- <(('-' / '+') timeHour ':' timeMinute)> */
		nil,
		/* 48 timeOffset <- <('Z' / timeNumoffset)> */
		nil,
		/* 49 partialTime <- <(timeHour ':' timeMinute ':' timeSecond timeSecfrac?)> */
		nil,
		/* 50 fullDate <- <(dateFullYear '-' dateMonth '-' dateMDay)> */
		nil,
		/* 51 fullTime <- <(partialTime timeOffset)> */
		nil,
		/* 52 datetime <- <(fullDate 'T' fullTime)> */
		nil,
		/* 53 digit <- <[0-9]> */
		func() bool {
			position287, tokenIndex287 := position, tokenIndex
			{
				position288 := position
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l287
				}
				position++
				add(ruledigit, position288)
			}
			return true
		l287:
			position, tokenIndex = position287, tokenIndex287
			return false
		},
		/* 54 digitDual <- <(digit digit)> */
		func() bool {
			position289, tokenIndex289 := position, tokenIndex
			{
				position290 := position
				if !_rules[ruledigit]() {
					goto l289
				}
				if !_rules[ruledigit]() {
					goto l289
				}
				add(ruledigitDual, position290)
			}
			return true
		l289:
			position, tokenIndex = position289, tokenIndex289
			return false
		},
		/* 55 digitQuad <- <(digitDual digitDual)> */
		nil,
		/* 56 array <- <('[' Action22 wsnl arrayValues wsnl ']')> */
		nil,
		/* 57 arrayValues <- <(val Action23 arraySep? (comment? newline)?)*> */
		nil,
		/* 58 arraySep <- <(ws ',' wsnl)> */
		nil,
		/* 60 Action0 <- <{ _ = buffer }> */
		nil,
		nil,
		/* 62 Action1 <- <{ p.SetTableString(begin, end) }> */
		nil,
		/* 63 Action2 <- <{ p.AddLineCount(end - begin) }> */
		nil,
		/* 64 Action3 <- <{ p.AddLineCount(end - begin) }> */
		nil,
		/* 65 Action4 <- <{ p.AddKeyValue() }> */
		nil,
		/* 66 Action5 <- <{ p.SetKey(p.buffer, begin, end) }> */
		nil,
		/* 67 Action6 <- <{ p.SetKey(p.buffer, begin-1, end+1) }> */
		nil,
		/* 68 Action7 <- <{ p.SetTime(begin, end) }> */
		nil,
		/* 69 Action8 <- <{ p.SetFloat64(begin, end) }> */
		nil,
		/* 70 Action9 <- <{ p.SetInt64(begin, end) }> */
		nil,
		/* 71 Action10 <- <{ p.SetString(begin, end) }> */
		nil,
		/* 72 Action11 <- <{ p.SetBool(begin, end) }> */
		nil,
		/* 73 Action12 <- <{ p.SetArray(begin, end) }> */
		nil,
		/* 74 Action13 <- <{ p.SetTable(p.buffer, begin, end) }> */
		nil,
		/* 75 Action14 <- <{ p.SetArrayTable(p.buffer, begin, end) }> */
		nil,
		/* 76 Action15 <- <{ p.StartInlineTable() }> */
		nil,
		/* 77 Action16 <- <{ p.EndInlineTable() }> */
		nil,
		/* 78 Action17 <- <{ p.SetBasicString(p.buffer, begin, end) }> */
		nil,
		/* 79 Action18 <- <{ p.SetMultilineString() }> */
		nil,
		/* 80 Action19 <- <{ p.AddMultilineBasicBody(p.buffer, begin, end) }> */
		nil,
		/* 81 Action20 <- <{ p.SetLiteralString(p.buffer, begin, end) }> */
		nil,
		/* 82 Action21 <- <{ p.SetMultilineLiteralString(p.buffer, begin, end) }> */
		nil,
		/* 83 Action22 <- <{ p.StartArray() }> */
		nil,
		/* 84 Action23 <- <{ p.AddArrayVal() }> */
		nil,
	}
	p.rules = _rules
}
