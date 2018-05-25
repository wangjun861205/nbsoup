package nbsoup

import (
	"errors"
	"io/ioutil"
	"regexp"
	"strings"
)

type queryOperator int

const (
	equal queryOperator = iota
	notEqual
	include
	except
	reg
)

var operatorMap = map[string]queryOperator{
	"=":   equal,
	"!=":  notEqual,
	"*=":  include,
	"!*=": except,
	"%=":  reg,
}

type queryRelation int

const (
	and queryRelation = iota
	or
)

var ErrEmptyQuery = errors.New("empty query string")
var ErrNoValidQuery = errors.New("no valid query in query string")
var ErrInvalidAttrName = errors.New("invalid attribute name")
var ErrInvalidOperator = errors.New("invalid operator")
var ErrInvalidCharacter = errors.New("invalid character")

var qRe = regexp.MustCompile(`(\w*)\[(.*?)\]`)

type query struct {
	name      string
	queryList [][]q
	next      *query
}

type q struct {
	name     string
	operator queryOperator
	value    string
}

func parseQuery(s string) (*query, error) {
	s = strings.Trim(s, " \t")
	if s == "" {
		return nil, ErrEmptyQuery
	}
	queryElem := ""
	reader := strings.NewReader(s)
	var inQuote bool
OUTER:
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			break
		}
		switch r {
		case '"':
			inQuote = !inQuote
			queryElem += string(r)
		case '.':
			if inQuote {
				queryElem += string(r)
				continue
			}
			break OUTER
		default:
			queryElem += string(r)
		}
	}
	if queryElem == "" {
		return nil, ErrNoValidQuery
	}
	// nameAndAttr := qRe.FindStringSubmatch(queryElem)
	// name := nameAndAttr[1]
	// attr := nameAndAttr[2]
	// qList, err := parseQ(attr)
	// if err != nil {
	// 	return nil, err
	// }
	subReader := strings.NewReader(queryElem)
	var name string
	var attrs string
	var inAttrs bool
SUB:
	for {
		r, _, err := subReader.ReadRune()
		if err != nil {
			break
		}
		switch r {
		case '[':
			inAttrs = true
		case ']':
			break SUB
		default:
			if inAttrs {
				attrs += string(r)
			} else {
				name += string(r)
			}
		}
	}
	qList, err := parseQ(attrs)
	if err != nil {
		return nil, err
	}
	thisQuery := query{name: name, queryList: qList}
	if reader.Len() > 0 {
		bRemain, _ := ioutil.ReadAll(reader)
		nextQuery, err := parseQuery(string(bRemain))
		if err != nil {
			return nil, err
		}
		thisQuery.next = nextQuery
	}
	return &thisQuery, nil
}

type attributePosition int

const (
	posName attributePosition = iota
	posOperator
	posValue
	posFinish
)

func parseQ(s string) ([][]q, error) {
	s = strings.Trim(s, " \t")
	if s == "" {
		return nil, nil
	}
	reader := strings.NewReader(s)
	queryList := make([][]q, 0, 16)
	var attrName string
	var operator string
	var value string
	var relation queryRelation
	var pos attributePosition
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			if ok := checkName(attrName); !ok {
				return nil, ErrInvalidAttrName
			}
			if ok := checkOperator(operator); !ok {
				return nil, ErrInvalidOperator
			}
			op := operatorMap[operator]
			if op == reg {
				_, err := regexp.Compile(value)
				if err != nil {
					return nil, err
				}
			}
			thisQ := q{name: attrName, operator: op, value: value}
			if len(queryList) == 0 {
				qList := []q{thisQ}
				queryList = append(queryList, qList)
			} else {
				if relation == and {
					index := len(queryList) - 1
					queryList[index] = append(queryList[index], thisQ)
				} else {
					qList := []q{thisQ}
					queryList = append(queryList, qList)
				}
			}
			return queryList, nil
		}
		switch r {
		case ' ':
			if pos == posValue {
				value += string(r)
			}
			continue
		case '=', '!', '*', '%':
			switch pos {
			case posName:
				pos = posOperator
				operator += string(r)
			case posOperator:
				operator += string(r)
			case posValue:
				value += string(r)
			case posFinish:
				return nil, ErrInvalidCharacter
			}
		case '"':
			switch pos {
			case posName:
				return nil, ErrInvalidAttrName
			case posOperator:
				pos = posValue
			case posValue:
				pos = posFinish
			case posFinish:
				return nil, ErrInvalidCharacter
			}
		case '&', '|':
			switch pos {
			case posName:
				return nil, ErrInvalidAttrName
			case posOperator:
				return nil, ErrInvalidOperator
			case posValue:
				value += string(r)
			case posFinish:
				if ok := checkName(attrName); !ok {
					return nil, ErrInvalidAttrName
				}
				if ok := checkOperator(operator); !ok {
					return nil, ErrInvalidOperator
				}
				op := operatorMap[operator]
				if op == reg {
					if _, err := regexp.Compile(value); err != nil {
						return nil, err
					}
				}
				thisQ := q{name: attrName, operator: op, value: value}
				if len(queryList) == 0 {
					qList := []q{thisQ}
					queryList = append(queryList, qList)
				} else {
					if relation == and {
						index := len(queryList) - 1
						queryList[index] = append(queryList[index], thisQ)
					} else {
						qList := []q{thisQ}
						queryList = append(queryList, qList)
					}
				}
				attrName = ""
				operator = ""
				value = ""
				pos = posName
				if string(r) == "&" {
					relation = and
				} else {
					relation = or
				}
			}
		default:
			switch pos {
			case posName:
				attrName += string(r)
			case posOperator:
				return nil, ErrInvalidOperator
			case posValue:
				value += string(r)
			case posFinish:
				return nil, ErrInvalidCharacter
			}
		}
	}
}

var nameCheckRe = regexp.MustCompile(`^\w+$`)

func checkName(attrName string) bool {
	return nameCheckRe.MatchString(attrName)
}

func checkOperator(op string) bool {
	_, ok := operatorMap[op]
	return ok
}
