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
								position94, tokenIndex94 := position, tokenIndex
								{
									position96 := position
									{
										position97 := position
										{
											position98 := position
											if !_rules[ruledigitDual]() {
												goto l95
											}
											if !_rules[ruledigitDual]() {
												goto l95
											}
											add(ruledigitQuad, position98)
										}
										add(ruledateFullYear, position97)
									}
									if buffer[position] != rune('-') {
										goto l95
									}
									position++
									{
										position99 := position
										if !_rules[ruledigitDual]() {
											goto l95
										}
										add(ruledateMonth, position99)
									}
									if buffer[position] != rune('-') {
										goto l95
									}
									position++
									{
										position100 := position
										if !_rules[ruledigitDual]() {
											goto l95
										}
										add(ruledateMDay, position100)
									}
									add(rulefullDate, position96)
								}
								{
									position101, tokenIndex101 := position, tokenIndex
									if buffer[position] != rune('T') {
										goto l101
									}
									position++
									{
										position103 := position
										if !_rules[rulepartialTime]() {
											goto l101
										}
										{
											position104 := position
											{
												position105, tokenIndex105 := position, tokenIndex
												if buffer[position] != rune('Z') {
													goto l106
												}
												position++
												goto l105
											l106:
												position, tokenIndex = position105, tokenIndex105
												{
													position107 := position
													{
														position108, tokenIndex108 := position, tokenIndex
														if buffer[position] != rune('-') {
															goto l109
														}
														position++
														goto l108
													l109:
														position, tokenIndex = position108, tokenIndex108
														if buffer[position] != rune('+') {
															goto l101
														}
														position++
													}
												l108:
													if !_rules[ruletimeHour]() {
														goto l101
													}
													if buffer[position] != rune(':') {
														goto l101
													}
													position++
													if !_rules[ruletimeMinute]() {
														goto l101
													}
													add(ruletimeNumoffset, position107)
												}
											}
										l105:
											add(ruletimeOffset, position104)
										}
										add(rulefullTime, position103)
									}
									goto l102
								l101:
									position, tokenIndex = position101, tokenIndex101
								}
							l102:
								goto l94
							l95:
								position, tokenIndex = position94, tokenIndex94
								if !_rules[rulepartialTime]() {
									goto l91
								}
							}
						l94:
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
						position112 := position
						{
							position113 := position
							if !_rules[ruleinteger]() {
								goto l111
							}
							{
								position114, tokenIndex114 := position, tokenIndex
								if !_rules[rulefrac]() {
									goto l115
								}
								{
									position116, tokenIndex116 := position, tokenIndex
									if !_rules[ruleexp]() {
										goto l116
									}
									goto l117
								l116:
									position, tokenIndex = position116, tokenIndex116
								}
							l117:
								goto l114
							l115:
								position, tokenIndex = position114, tokenIndex114
								{
									position118, tokenIndex118 := position, tokenIndex
									if !_rules[rulefrac]() {
										goto l118
									}
									goto l119
								l118:
									position, tokenIndex = position118, tokenIndex118
								}
							l119:
								if !_rules[ruleexp]() {
									goto l111
								}
							}
						l114:
							add(rulefloat, position113)
						}
						add(rulePegText, position112)
					}
					{
						add(ruleAction8, position)
					}
					goto l90
				l111:
					position, tokenIndex = position90, tokenIndex90
					{
						switch buffer[position] {
						case '{':
							{
								position122 := position
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
									position124 := position
								l125:
									{
										position126, tokenIndex126 := position, tokenIndex
										if !_rules[rulekeyval]() {
											goto l126
										}
										{
											position127, tokenIndex127 := position, tokenIndex
											{
												position129 := position
												if !_rules[rulews]() {
													goto l127
												}
												if buffer[position] != rune(',') {
													goto l127
												}
												position++
												if !_rules[rulews]() {
													goto l127
												}
												add(ruleinlineTableValSep, position129)
											}
											goto l128
										l127:
											position, tokenIndex = position127, tokenIndex127
										}
									l128:
										goto l125
									l126:
										position, tokenIndex = position126, tokenIndex126
									}
									add(ruleinlineTableKeyValues, position124)
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
								add(ruleinlineTable, position122)
							}
							break
						case '[':
							{
								position131 := position
								{
									position132 := position
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
										position134 := position
									l135:
										{
											position136, tokenIndex136 := position, tokenIndex
											if !_rules[ruleval]() {
												goto l136
											}
											{
												add(ruleAction23, position)
											}
											{
												position138, tokenIndex138 := position, tokenIndex
												{
													position140 := position
													if !_rules[rulews]() {
														goto l138
													}
													if buffer[position] != rune(',') {
														goto l138
													}
													position++
													if !_rules[rulewsnl]() {
														goto l138
													}
													add(rulearraySep, position140)
												}
												goto l139
											l138:
												position, tokenIndex = position138, tokenIndex138
											}
										l139:
											{
												position141, tokenIndex141 := position, tokenIndex
												{
													position143, tokenIndex143 := position, tokenIndex
													if !_rules[rulecomment]() {
														goto l143
													}
													goto l144
												l143:
													position, tokenIndex = position143, tokenIndex143
												}
											l144:
												if !_rules[rulenewline]() {
													goto l141
												}
												goto l142
											l141:
												position, tokenIndex = position141, tokenIndex141
											}
										l142:
											goto l135
										l136:
											position, tokenIndex = position136, tokenIndex136
										}
										add(rulearrayValues, position134)
									}
									if !_rules[rulewsnl]() {
										goto l88
									}
									if buffer[position] != rune(']') {
										goto l88
									}
									position++
									add(rulearray, position132)
								}
								add(rulePegText, position131)
							}
							{
								add(ruleAction12, position)
							}
							break
						case 'f', 't':
							{
								position146 := position
								{
									position147 := position
									{
										position148, tokenIndex148 := position, tokenIndex
										if buffer[position] != rune('t') {
											goto l149
										}
										position++
										if buffer[position] != rune('r') {
											goto l149
										}
										position++
										if buffer[position] != rune('u') {
											goto l149
										}
										position++
										if buffer[position] != rune('e') {
											goto l149
										}
										position++
										goto l148
									l149:
										position, tokenIndex = position148, tokenIndex148
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
								l148:
									add(ruleboolean, position147)
								}
								add(rulePegText, position146)
							}
							{
								add(ruleAction11, position)
							}
							break
						case '"', '\'':
							{
								position151 := position
								{
									position152 := position
									{
										position153, tokenIndex153 := position, tokenIndex
										{
											position155 := position
											if buffer[position] != rune('\'') {
												goto l154
											}
											position++
											if buffer[position] != rune('\'') {
												goto l154
											}
											position++
											if buffer[position] != rune('\'') {
												goto l154
											}
											position++
											{
												position156 := position
												{
													position157 := position
												l158:
													{
														position159, tokenIndex159 := position, tokenIndex
														{
															position160, tokenIndex160 := position, tokenIndex
															if buffer[position] != rune('\'') {
																goto l160
															}
															position++
															if buffer[position] != rune('\'') {
																goto l160
															}
															position++
															if buffer[position] != rune('\'') {
																goto l160
															}
															position++
															goto l159
														l160:
															position, tokenIndex = position160, tokenIndex160
														}
														{
															position161, tokenIndex161 := position, tokenIndex
															{
																position163 := position
																{
																	position164, tokenIndex164 := position, tokenIndex
																	if buffer[position] != rune('\t') {
																		goto l165
																	}
																	position++
																	goto l164
																l165:
																	position, tokenIndex = position164, tokenIndex164
																	if c := buffer[position]; c < rune(' ') || c > rune('\U0010ffff') {
																		goto l162
																	}
																	position++
																}
															l164:
																add(rulemlLiteralChar, position163)
															}
															goto l161
														l162:
															position, tokenIndex = position161, tokenIndex161
															if !_rules[rulenewline]() {
																goto l159
															}
														}
													l161:
														goto l158
													l159:
														position, tokenIndex = position159, tokenIndex159
													}
													add(rulemlLiteralBody, position157)
												}
												add(rulePegText, position156)
											}
											if buffer[position] != rune('\'') {
												goto l154
											}
											position++
											if buffer[position] != rune('\'') {
												goto l154
											}
											position++
											if buffer[position] != rune('\'') {
												goto l154
											}
											position++
											{
												add(ruleAction21, position)
											}
											add(rulemlLiteralString, position155)
										}
										goto l153
									l154:
										position, tokenIndex = position153, tokenIndex153
										{
											position168 := position
											if buffer[position] != rune('\'') {
												goto l167
											}
											position++
											{
												position169 := position
											l170:
												{
													position171, tokenIndex171 := position, tokenIndex
													{
														position172 := position
														{
															switch buffer[position] {
															case '\t':
																if buffer[position] != rune('\t') {
																	goto l171
																}
																position++
																break
															case ' ', '!', '"', '#', '$', '%', '&':
																if c := buffer[position]; c < rune(' ') || c > rune('&') {
																	goto l171
																}
																position++
																break
															default:
																if c := buffer[position]; c < rune('(') || c > rune('\U0010ffff') {
																	goto l171
																}
																position++
																break
															}
														}

														add(ruleliteralChar, position172)
													}
													goto l170
												l171:
													position, tokenIndex = position171, tokenIndex171
												}
												add(rulePegText, position169)
											}
											if buffer[position] != rune('\'') {
												goto l167
											}
											position++
											{
												add(ruleAction20, position)
											}
											add(ruleliteralString, position168)
										}
										goto l153
									l167:
										position, tokenIndex = position153, tokenIndex153
										{
											position176 := position
											if buffer[position] != rune('"') {
												goto l175
											}
											position++
											if buffer[position] != rune('"') {
												goto l175
											}
											position++
											if buffer[position] != rune('"') {
												goto l175
											}
											position++
											{
												position177 := position
											l178:
												{
													position179, tokenIndex179 := position, tokenIndex
													{
														position180, tokenIndex180 := position, tokenIndex
														{
															position182 := position
															{
																position183, tokenIndex183 := position, tokenIndex
																if !_rules[rulebasicChar]() {
																	goto l184
																}
																goto l183
															l184:
																position, tokenIndex = position183, tokenIndex183
																if !_rules[rulenewline]() {
																	goto l181
																}
															}
														l183:
															add(rulePegText, position182)
														}
														{
															add(ruleAction19, position)
														}
														goto l180
													l181:
														position, tokenIndex = position180, tokenIndex180
														if !_rules[ruleescape]() {
															goto l179
														}
														if !_rules[rulenewline]() {
															goto l179
														}
														if !_rules[rulewsnl]() {
															goto l179
														}
													}
												l180:
													goto l178
												l179:
													position, tokenIndex = position179, tokenIndex179
												}
												add(rulemlBasicBody, position177)
											}
											if buffer[position] != rune('"') {
												goto l175
											}
											position++
											if buffer[position] != rune('"') {
												goto l175
											}
											position++
											if buffer[position] != rune('"') {
												goto l175
											}
											position++
											{
												add(ruleAction18, position)
											}
											add(rulemlBasicString, position176)
										}
										goto l153
									l175:
										position, tokenIndex = position153, tokenIndex153
										{
											position187 := position
											{
												position188 := position
												if buffer[position] != rune('"') {
													goto l88
												}
												position++
											l189:
												{
													position190, tokenIndex190 := position, tokenIndex
													if !_rules[rulebasicChar]() {
														goto l190
													}
													goto l189
												l190:
													position, tokenIndex = position190, tokenIndex190
												}
												if buffer[position] != rune('"') {
													goto l88
												}
												position++
												add(rulePegText, position188)
											}
											{
												add(ruleAction17, position)
											}
											add(rulebasicString, position187)
										}
									}
								l153:
									add(rulestring, position152)
								}
								add(rulePegText, position151)
							}
							{
								add(ruleAction10, position)
							}
							break
						default:
							{
								position193 := position
								if !_rules[ruleinteger]() {
									goto l88
								}
								add(rulePegText, position193)
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
			position200, tokenIndex200 := position, tokenIndex
			{
				position201 := position
				if !_rules[rulekey]() {
					goto l200
				}
			l202:
				{
					position203, tokenIndex203 := position, tokenIndex
					{
						position204 := position
						if !_rules[rulews]() {
							goto l203
						}
						if buffer[position] != rune('.') {
							goto l203
						}
						position++
						if !_rules[rulews]() {
							goto l203
						}
						add(ruletableKeySep, position204)
					}
					if !_rules[rulekey]() {
						goto l203
					}
					goto l202
				l203:
					position, tokenIndex = position203, tokenIndex203
				}
				add(ruletableKey, position201)
			}
			return true
		l200:
			position, tokenIndex = position200, tokenIndex200
			return false
		},
		/* 17 tableKeySep <- <(ws '.' ws)> */
		nil,
		/* 18 inlineTableValSep <- <(ws ',' ws)> */
		nil,
		/* 19 integer <- <(('-' / '+')? int)> */
		func() bool {
			position207, tokenIndex207 := position, tokenIndex
			{
				position208 := position
				{
					position209, tokenIndex209 := position, tokenIndex
					{
						position211, tokenIndex211 := position, tokenIndex
						if buffer[position] != rune('-') {
							goto l212
						}
						position++
						goto l211
					l212:
						position, tokenIndex = position211, tokenIndex211
						if buffer[position] != rune('+') {
							goto l209
						}
						position++
					}
				l211:
					goto l210
				l209:
					position, tokenIndex = position209, tokenIndex209
				}
			l210:
				{
					position213 := position
					{
						position214, tokenIndex214 := position, tokenIndex
						if c := buffer[position]; c < rune('1') || c > rune('9') {
							goto l215
						}
						position++
						{
							position218, tokenIndex218 := position, tokenIndex
							if !_rules[ruledigit]() {
								goto l219
							}
							goto l218
						l219:
							position, tokenIndex = position218, tokenIndex218
							if buffer[position] != rune('_') {
								goto l215
							}
							position++
							if !_rules[ruledigit]() {
								goto l215
							}
						}
					l218:
					l216:
						{
							position217, tokenIndex217 := position, tokenIndex
							{
								position220, tokenIndex220 := position, tokenIndex
								if !_rules[ruledigit]() {
									goto l221
								}
								goto l220
							l221:
								position, tokenIndex = position220, tokenIndex220
								if buffer[position] != rune('_') {
									goto l217
								}
								position++
								if !_rules[ruledigit]() {
									goto l217
								}
							}
						l220:
							goto l216
						l217:
							position, tokenIndex = position217, tokenIndex217
						}
						goto l214
					l215:
						position, tokenIndex = position214, tokenIndex214
						if !_rules[ruledigit]() {
							goto l207
						}
					}
				l214:
					add(ruleint, position213)
				}
				add(ruleinteger, position208)
			}
			return true
		l207:
			position, tokenIndex = position207, tokenIndex207
			return false
		},
		/* 20 int <- <(([1-9] (digit / ('_' digit))+) / digit)> */
		nil,
		/* 21 float <- <(integer ((frac exp?) / (frac? exp)))> */
		nil,
		/* 22 frac <- <('.' digit (digit / ('_' digit))*)> */
		func() bool {
			position224, tokenIndex224 := position, tokenIndex
			{
				position225 := position
				if buffer[position] != rune('.') {
					goto l224
				}
				position++
				if !_rules[ruledigit]() {
					goto l224
				}
			l226:
				{
					position227, tokenIndex227 := position, tokenIndex
					{
						position228, tokenIndex228 := position, tokenIndex
						if !_rules[ruledigit]() {
							goto l229
						}
						goto l228
					l229:
						position, tokenIndex = position228, tokenIndex228
						if buffer[position] != rune('_') {
							goto l227
						}
						position++
						if !_rules[ruledigit]() {
							goto l227
						}
					}
				l228:
					goto l226
				l227:
					position, tokenIndex = position227, tokenIndex227
				}
				add(rulefrac, position225)
			}
			return true
		l224:
			position, tokenIndex = position224, tokenIndex224
			return false
		},
		/* 23 exp <- <(('e' / 'E') ('-' / '+')? digit (digit / ('_' digit))*)> */
		func() bool {
			position230, tokenIndex230 := position, tokenIndex
			{
				position231 := position
				{
					position232, tokenIndex232 := position, tokenIndex
					if buffer[position] != rune('e') {
						goto l233
					}
					position++
					goto l232
				l233:
					position, tokenIndex = position232, tokenIndex232
					if buffer[position] != rune('E') {
						goto l230
					}
					position++
				}
			l232:
				{
					position234, tokenIndex234 := position, tokenIndex
					{
						position236, tokenIndex236 := position, tokenIndex
						if buffer[position] != rune('-') {
							goto l237
						}
						position++
						goto l236
					l237:
						position, tokenIndex = position236, tokenIndex236
						if buffer[position] != rune('+') {
							goto l234
						}
						position++
					}
				l236:
					goto l235
				l234:
					position, tokenIndex = position234, tokenIndex234
				}
			l235:
				if !_rules[ruledigit]() {
					goto l230
				}
			l238:
				{
					position239, tokenIndex239 := position, tokenIndex
					{
						position240, tokenIndex240 := position, tokenIndex
						if !_rules[ruledigit]() {
							goto l241
						}
						goto l240
					l241:
						position, tokenIndex = position240, tokenIndex240
						if buffer[position] != rune('_') {
							goto l239
						}
						position++
						if !_rules[ruledigit]() {
							goto l239
						}
					}
				l240:
					goto l238
				l239:
					position, tokenIndex = position239, tokenIndex239
				}
				add(ruleexp, position231)
			}
			return true
		l230:
			position, tokenIndex = position230, tokenIndex230
			return false
		},
		/* 24 string <- <(mlLiteralString / literalString / mlBasicString / basicString)> */
		nil,
		/* 25 basicString <- <(<('"' basicChar* '"')> Action17)> */
		nil,
		/* 26 basicChar <- <(basicUnescaped / escaped)> */
		func() bool {
			position244, tokenIndex244 := position, tokenIndex
			{
				position245 := position
				{
					position246, tokenIndex246 := position, tokenIndex
					{
						position248 := position
						{
							switch buffer[position] {
							case ' ', '!':
								if c := buffer[position]; c < rune(' ') || c > rune('!') {
									goto l247
								}
								position++
								break
							case '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ':', ';', '<', '=', '>', '?', '@', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '[':
								if c := buffer[position]; c < rune('#') || c > rune('[') {
									goto l247
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune(']') || c > rune('\U0010ffff') {
									goto l247
								}
								position++
								break
							}
						}

						add(rulebasicUnescaped, position248)
					}
					goto l246
				l247:
					position, tokenIndex = position246, tokenIndex246
					{
						position250 := position
						if !_rules[ruleescape]() {
							goto l244
						}
						{
							switch buffer[position] {
							case 'U':
								if buffer[position] != rune('U') {
									goto l244
								}
								position++
								if !_rules[rulehexQuad]() {
									goto l244
								}
								if !_rules[rulehexQuad]() {
									goto l244
								}
								break
							case 'u':
								if buffer[position] != rune('u') {
									goto l244
								}
								position++
								if !_rules[rulehexQuad]() {
									goto l244
								}
								break
							case '\\':
								if buffer[position] != rune('\\') {
									goto l244
								}
								position++
								break
							case '/':
								if buffer[position] != rune('/') {
									goto l244
								}
								position++
								break
							case '"':
								if buffer[position] != rune('"') {
									goto l244
								}
								position++
								break
							case 'r':
								if buffer[position] != rune('r') {
									goto l244
								}
								position++
								break
							case 'f':
								if buffer[position] != rune('f') {
									goto l244
								}
								position++
								break
							case 'n':
								if buffer[position] != rune('n') {
									goto l244
								}
								position++
								break
							case 't':
								if buffer[position] != rune('t') {
									goto l244
								}
								position++
								break
							default:
								if buffer[position] != rune('b') {
									goto l244
								}
								position++
								break
							}
						}

						add(ruleescaped, position250)
					}
				}
			l246:
				add(rulebasicChar, position245)
			}
			return true
		l244:
			position, tokenIndex = position244, tokenIndex244
			return false
		},
		/* 27 escaped <- <(escape ((&('U') ('U' hexQuad hexQuad)) | (&('u') ('u' hexQuad)) | (&('\\') '\\') | (&('/') '/') | (&('"') '"') | (&('r') 'r') | (&('f') 'f') | (&('n') 'n') | (&('t') 't') | (&('b') 'b')))> */
		nil,
		/* 28 basicUnescaped <- <((&(' ' | '!') [ -!]) | (&('#' | '$' | '%' | '&' | '\'' | '(' | ')' | '*' | '+' | ',' | '-' | '.' | '/' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | ':' | ';' | '<' | '=' | '>' | '?' | '@' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '[') [#-[]) | (&(']' | '^' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z' | '{' | '|' | '}' | '~' | '\u007f' | '\u0080' | '\u0081' | '\u0082' | '\u0083' | '\u0084' | '\u0085' | '\u0086' | '\u0087' | '\u0088' | '\u0089' | '\u008a' | '\u008b' | '\u008c' | '\u008d' | '\u008e' | '\u008f' | '\u0090' | '\u0091' | '\u0092' | '\u0093' | '\u0094' | '\u0095' | '\u0096' | '\u0097' | '\u0098' | '\u0099' | '\u009a' | '\u009b' | '\u009c' | '\u009d' | '\u009e' | '\u009f' | '\u00a0' | '¡' | '¢' | '£' | '¤' | '¥' | '¦' | '§' | '¨' | '©' | 'ª' | '«' | '¬' | '\u00ad' | '®' | '¯' | '°' | '±' | '²' | '³' | '´' | 'µ' | '¶' | '·' | '¸' | '¹' | 'º' | '»' | '¼' | '½' | '¾' | '¿' | 'À' | 'Á' | 'Â' | 'Ã' | 'Ä' | 'Å' | 'Æ' | 'Ç' | 'È' | 'É' | 'Ê' | 'Ë' | 'Ì' | 'Í' | 'Î' | 'Ï' | 'Ð' | 'Ñ' | 'Ò' | 'Ó' | 'Ô' | 'Õ' | 'Ö' | '×' | 'Ø' | 'Ù' | 'Ú' | 'Û' | 'Ü' | 'Ý' | 'Þ' | 'ß' | 'à' | 'á' | 'â' | 'ã' | 'ä' | 'å' | 'æ' | 'ç' | 'è' | 'é' | 'ê' | 'ë' | 'ì' | 'í' | 'î' | 'ï' | 'ð' | 'ñ' | 'ò' | 'ó' | 'ô' | 'õ' | 'ö' | '÷' | 'ø' | 'ù' | 'ú' | 'û' | 'ü' | 'ý' | 'þ' | 'ÿ') []-\U0010ffff]))> */
		nil,
		/* 29 escape <- <'\\'> */
		func() bool {
			position254, tokenIndex254 := position, tokenIndex
			{
				position255 := position
				if buffer[position] != rune('\\') {
					goto l254
				}
				position++
				add(ruleescape, position255)
			}
			return true
		l254:
			position, tokenIndex = position254, tokenIndex254
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
			position263, tokenIndex263 := position, tokenIndex
			{
				position264 := position
				{
					switch buffer[position] {
					case 'a', 'b', 'c', 'd', 'e', 'f':
						if c := buffer[position]; c < rune('a') || c > rune('f') {
							goto l263
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F':
						if c := buffer[position]; c < rune('A') || c > rune('F') {
							goto l263
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l263
						}
						position++
						break
					}
				}

				add(rulehexdigit, position264)
			}
			return true
		l263:
			position, tokenIndex = position263, tokenIndex263
			return false
		},
		/* 38 hexQuad <- <(hexdigit hexdigit hexdigit hexdigit)> */
		func() bool {
			position266, tokenIndex266 := position, tokenIndex
			{
				position267 := position
				if !_rules[rulehexdigit]() {
					goto l266
				}
				if !_rules[rulehexdigit]() {
					goto l266
				}
				if !_rules[rulehexdigit]() {
					goto l266
				}
				if !_rules[rulehexdigit]() {
					goto l266
				}
				add(rulehexQuad, position267)
			}
			return true
		l266:
			position, tokenIndex = position266, tokenIndex266
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
			position272, tokenIndex272 := position, tokenIndex
			{
				position273 := position
				if !_rules[ruledigitDual]() {
					goto l272
				}
				add(ruletimeHour, position273)
			}
			return true
		l272:
			position, tokenIndex = position272, tokenIndex272
			return false
		},
		/* 44 timeMinute <- <digitDual> */
		func() bool {
			position274, tokenIndex274 := position, tokenIndex
			{
				position275 := position
				if !_rules[ruledigitDual]() {
					goto l274
				}
				add(ruletimeMinute, position275)
			}
			return true
		l274:
			position, tokenIndex = position274, tokenIndex274
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
		func() bool {
			position280, tokenIndex280 := position, tokenIndex
			{
				position281 := position
				if !_rules[ruletimeHour]() {
					goto l280
				}
				if buffer[position] != rune(':') {
					goto l280
				}
				position++
				if !_rules[ruletimeMinute]() {
					goto l280
				}
				if buffer[position] != rune(':') {
					goto l280
				}
				position++
				{
					position282 := position
					if !_rules[ruledigitDual]() {
						goto l280
					}
					add(ruletimeSecond, position282)
				}
				{
					position283, tokenIndex283 := position, tokenIndex
					{
						position285 := position
						if buffer[position] != rune('.') {
							goto l283
						}
						position++
						if !_rules[ruledigit]() {
							goto l283
						}
					l286:
						{
							position287, tokenIndex287 := position, tokenIndex
							if !_rules[ruledigit]() {
								goto l287
							}
							goto l286
						l287:
							position, tokenIndex = position287, tokenIndex287
						}
						add(ruletimeSecfrac, position285)
					}
					goto l284
				l283:
					position, tokenIndex = position283, tokenIndex283
				}
			l284:
				add(rulepartialTime, position281)
			}
			return true
		l280:
			position, tokenIndex = position280, tokenIndex280
			return false
		},
		/* 50 fullDate <- <(dateFullYear '-' dateMonth '-' dateMDay)> */
		nil,
		/* 51 fullTime <- <(partialTime timeOffset)> */
		nil,
		/* 52 datetime <- <((fullDate ('T' fullTime)?) / partialTime)> */
		nil,
		/* 53 digit <- <[0-9]> */
		func() bool {
			position291, tokenIndex291 := position, tokenIndex
			{
				position292 := position
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l291
				}
				position++
				add(ruledigit, position292)
			}
			return true
		l291:
			position, tokenIndex = position291, tokenIndex291
			return false
		},
		/* 54 digitDual <- <(digit digit)> */
		func() bool {
			position293, tokenIndex293 := position, tokenIndex
			{
				position294 := position
				if !_rules[ruledigit]() {
					goto l293
				}
				if !_rules[ruledigit]() {
					goto l293
				}
				add(ruledigitDual, position294)
			}
			return true
		l293:
			position, tokenIndex = position293, tokenIndex293
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
