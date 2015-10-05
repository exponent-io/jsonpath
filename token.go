package json

import (
	"bytes"
	"encoding/json"
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

func (d *Decoder) pushNav(n interface{}) {
	d.navStack = append(d.navStack, n)
}

func (d *Decoder) popNav() interface{} {
	last := len(d.navStack) - 1
	n := d.navStack[last]

	if o, ok := n.(int); ok {
		d.arrayOffset = o + 1
	} else {
		d.arrayOffset = 0
	}
	d.navStack = d.navStack[:last]
	return n
}

func (d *Decoder) Path() JsonPath {
	p := make(JsonPath, len(d.path))
	copy(p, d.path)
	return p
}

// Token is basically equivalent to the Token() method on json.Decoder. The primary difference is that it distinguishes
// between strings that are keys and values. String tokens that are object keys are returned as the KeyString	type
// rather than as a bare string type.
func (d *Decoder) Token() (json.Token, error) {
	t, err := d.Decoder.Token()
	if err != nil {
		return t, err
	}

	if t == nil {
		return t, err
	}
	switch t := t.(type) {
	case json.Delim:
		switch t {
		case json.Delim('{'):
			if d.context == ArrayValue {
				d.path.inc()
			}
			d.path.openObj()
			d.context = ObjectKey
			break
		case json.Delim('}'):
			d.path.closeObj()
			d.context = d.path.inferContext()
			break
		case json.Delim('['):
			if d.context == ArrayValue {
				d.path.inc()
			}
			d.path.openArr()
			d.context = ArrayValue
			break
		case json.Delim(']'):
			d.path.closeArr()
			d.context = d.path.inferContext()
			break
		}
	case float64, json.Number, bool:
		switch d.context {
		case ObjectValue:
			d.context = ObjectKey
			break
		case ArrayValue:
			d.path.inc()
			break
		}
		break
	case string:
		switch d.context {
		case ObjectKey:
			d.path.name(t)
			d.context = ObjectValue
			return KeyString(t), err
		case ObjectValue:
			d.context = ObjectKey
		case ArrayValue:
			d.path.inc()
		}
		break
	}

	return t, err
}
