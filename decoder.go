package json

import (
	"encoding/json"
	"io"
)

// SeekingDecoder extends the encoding/json.Decoder to support seeking to a position
// in a JSON token stream using the SeekTo() method.
type Decoder struct {
	json.Decoder

	path    JsonPath
	context JsonContext
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{Decoder: *json.NewDecoder(r)}
}

// SeekTo causes the Decoder to move forward to a given path in the JSON structure.
//
// The path argument must consist of strings or integers. Each string specifies an JSON object key, and
// each integer specifies an index into a JSON array.
//
// Consider the JSON structure
//
//  { "a": [0,"s",12e4,{"b":0,"v":35} ] }
//
// SeekTo("a",3,"v") will move to the value referenced by the "a" key in the current object,
// followed by a move to the 4th value (index 3) in the array, followed by a move to the value at key "v".
// In this example, a subsequent call to the decoder's Decode() would unmarshal the value 35.
//
// SeekTo returns a boolean value indicating whether a match was found.
//
// Decoder is intended to be used with a stream of tokens. As a result it navigates forward only.
//
// The Decoder also keeps state about its position in the token stream. It is safe to use the Decode()
// method to read JSON values between calls to SeekTo. However, calling Token() between calls
// to SeekTo could confuse the SeekingDecoder unless care is made to always read all the tokens that comprise a
// JSON value before calling SeekTo again.
func (w *Decoder) SeekTo(path ...interface{}) (bool, error) {

	if len(path) == 0 {
		return len(w.path) == 0, nil
	}
	last := len(path) - 1
	if i, ok := path[last].(int); ok {
		path[last] = i - 1
	}

	for {
		if w.path.Equal(path) {
			return true, nil
		}
		_, err := w.Token()
		if err == io.EOF {
			return false, nil
		} else if err != nil {
			return false, err
		}
	}
}

func (d *Decoder) Decode(v interface{}) error {
	switch d.context {
	case ObjectValue:
		d.context = ObjectKey
		break
	case ArrayValue:
		d.path.inc()
		break
	}
	return d.Decoder.Decode(v)
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
