package json

import (
	"bytes"
	"fmt"
)

type KeyString string

type JsonContext int

const (
	None JsonContext = iota
	ObjectKey
	ObjectValue
	ArrayValue
)

type JsonPath []interface{}

func (p *JsonPath) openObj()           { p.push("") }
func (p *JsonPath) closeObj()          { p.pop() }
func (p *JsonPath) openArr()           { p.push(-1) }
func (p *JsonPath) closeArr()          { p.pop() }
func (p *JsonPath) push(n interface{}) { *p = append(*p, n) }
func (p *JsonPath) pop()               { *p = (*p)[:len(*p)-1] }
func (p *JsonPath) top() interface{}   { return (*p)[len(*p)-1] }
func (p *JsonPath) inc()               { (*p)[len(*p)-1] = (*p)[len(*p)-1].(int) + 1 }
func (p *JsonPath) name(n string)      { (*p)[len(*p)-1] = n }

func (p *JsonPath) inferContext() JsonContext {
	if len(*p) == 0 {
		return None
	}
	t := p.top()
	switch t.(type) {
	case string:
		return ObjectKey
	case int:
		return ArrayValue
	default:
		panic(fmt.Sprintf("Invalid stack type %T", t))
	}
}

func (p *JsonPath) Equal(o JsonPath) bool {
	if len(*p) != len(o) {
		return false
	}
	for i, v := range *p {
		if v != o[i] {
			return false
		}
	}
	return true
}

func (p *JsonPath) String() string {
	buff := &bytes.Buffer{}
	for _, v := range *p {
		switch v := v.(type) {
		case string:
			fmt.Fprintf(buff, "%q", v)
			break
		case int:
			fmt.Fprintf(buff, "%d", v)
			break
		}
		buff.WriteRune('.')
	}
	buff.Truncate(buff.Len() - 1)
	return buff.String()
}
