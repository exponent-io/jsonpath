package json_walk

import (
	"encoding/json"
	"fmt"
	"io"
)

type Walker struct {
	dec *json.Decoder

	navStack    []interface{} // keeps track of the move operations so we can figure out where the next one should be
	arrayOffset int           // saves the array offset so we can continue MoveToIndex where we left off
}

func NewWalker(dec *json.Decoder) *Walker {
	return &Walker{dec: dec}
}

// MoveTo causes the Decoder to move forward to a position in the JSON structure.
//
// The path argument must consist of strings or integers. Each string specifies an JSON object key, and
// each integer specifies an index into a JSON array.
//
// Consider the JSON structure
//
//  { "a": [0,"s",12e4,{"b":0,"v":35} ] }
//
// MoveTo("a",3,"v") will move to the value referenced by the "a" key in the current object,
// followed by a move to the 4th value (index 3) in the array, followed by a move to the value at key "v".
// In this example, a subsequent call to Next or Decode would unmarshal the value 35.
//
// MoveTo returns a boolean value indicating whether a match was found.
func (w *Walker) MoveTo(path ...interface{}) (bool, error) {

	var matched bool
	var err error

	// if we've already moved before, advance to the next token that shares a common
	// prefix with the current location
	if len(w.navStack) != 0 {
		matched, err = w.moveToCommonPrefix(path...)
		if !matched || err != nil {
			return matched, err
		}
		path = path[len(w.navStack):]
	}

	for _, p := range path {

		switch p := p.(type) {
		case int:
			matched, err = w.moveToIndex(p)
			if !matched || err != nil {
				return matched, err
			}
		case string:
			matched, err = w.moveToKey(p)
			if !matched || err != nil {
				return matched, err
			}
		default:
			return false, fmt.Errorf("invalid JSON path item '%v', must be a string or an int", p)
		}
	}
	return true, nil
}

// moveToKey traverses to the JSON value corresponding to the provided key.
// The decoder must be currently positioned on a JSON object. MoveToKey returns a boolean value
// indicating whether the key was found as well as an error value if an error occurred while traversing
// the JSON structure.
func (w *Walker) moveToKey(s string) (bool, error) {

	fmt.Printf("\nmoveToKey: %q\n", s)

	var st json.Token
	var err error
	var depth = 0 // we start at the beginning of an object

	for {
		st, err = w.dec.Token()
		fmt.Printf("mTK token: %T %v, %v\n", st, st, err)
		if err == io.EOF {
			return false, nil
		} else if err != nil {
			return false, err
		}

		switch st := st.(type) {
		case string:
			fmt.Printf("string %v %v %v\n", depth, s, st)
			if depth <= 1 && s == st {
				w.pushNav(s)
				return true, nil
			}
		case json.Delim:
			fmt.Printf("delim %v %v %v\n", depth, s, st)
			switch st {
			case json.Delim('{'):
				depth++
			case json.Delim('}'):
				depth--
				if depth <= 0 {
					return false, nil
				}
			}
		default:
			fmt.Printf("default %v %v %v\n", depth, s, st)
		}
	}
	return false, err
}

// moveToIndex traverses to the JSON value corresponding to the provided array offset.
// The decoder must be currently positioned on a JSON array. MoveToIndex returns a boolean value
// indicating whether the value was found as well as an error value if an error occurred while traversing
// the JSON structure.
func (w *Walker) moveToIndex(n int) (bool, error) {

	var err error
	var st json.Token
	var depth = 0

	skipped := w.arrayOffset

	fmt.Printf("\nmoveToIndex: %v\n skipped: %v", n, skipped)

	for {
		st, err = w.dec.Token()
		fmt.Printf("mTI token: %T %v, %v\n", st, st, err)
		if err == io.EOF {
			return false, nil
		} else if err != nil {
			return false, err
		}

		switch st {
		case json.Delim('['):
			depth++
		case json.Delim(']'):
			depth--
			if depth == 0 {
				return false, nil
			}
		case json.Delim('{'):
			depth++
		case json.Delim('}'):
			depth--
		}
		fmt.Printf("token %v %v %v\n", depth, n, st)

		if depth == 1 {
			skipped++
			if skipped == n+1 {
				w.pushNav(n)
				return true, nil
			}
		}
	}
	return false, err
}

// moveToCommonPrefix moves the decoder to a point in the JSON structure that shares a
// common prefix with the last MoveTo operation. This allows for repeated calls to MoveTo
// without the caller having to worry about translating absolute moves into relative moves.
func (w *Walker) moveToCommonPrefix(path ...interface{}) (bool, error) {

	fmt.Printf("\nmoveToCommonPrefix: %v\n", path)

	w.popNav()

	targetDepth := w.countCommonPrefix(path...)

	var err error
	var st json.Token
	var depth = len(w.navStack)

	fmt.Printf("depth %v %v\n", depth, w.navStack)

	if depth == targetDepth {
		return true, nil
	}

	for {
		st, err = w.dec.Token()
		fmt.Printf("mTCP token: %T %v, %v\n", st, st, err)
		if err == io.EOF {
			return false, nil
		} else if err != nil {
			return false, err
		}

		switch st {
		case json.Delim('['):
			depth++
		case json.Delim(']'):
			depth--
		case json.Delim('{'):
			depth++
		case json.Delim('}'):
			depth--
		}
		fmt.Printf("token %v %v\n", depth, st)

		if depth == targetDepth {
			w.setNavLen(depth)
			return true, nil
		}
	}
	return false, err
}

func (w *Walker) pushNav(n interface{}) {
	w.navStack = append(w.navStack, n)
}

func (w *Walker) popNav() interface{} {
	last := len(w.navStack) - 1
	n := w.navStack[last]

	if o, ok := n.(int); ok {
		w.arrayOffset = o + 1
	} else {
		w.arrayOffset = 0
	}
	w.navStack = w.navStack[:last]
	return n
}

func (w *Walker) setNavLen(depth int) {
	n := len(w.navStack) - depth
	for i := 0; i < n; i++ {
		w.popNav()
	}
}

func (w *Walker) countCommonPrefix(path ...interface{}) int {
	for i, p := range path {
		if i == len(w.navStack) || p != w.navStack[i] {
			return i
		}
	}
	return len(path)
}
