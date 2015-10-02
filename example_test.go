package json_walk

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func ExampleDecoder_MoveTo() {

	var j = []byte(`[
		{"Space": "YCbCr", "Point": {"Y": 255, "Cb": 0, "Cr": -10}},
		{"Space": "RGB",   "Point": {"R": 98, "G": 218, "B": 255}}
	]`)

	d := json.NewDecoder(bytes.NewReader(j))
	w := NewWalker(d)
	var v interface{}

	w.MoveTo(0, "Space")
	d.Decode(&v)
	fmt.Println(0, "Space", v)

	w.MoveTo(0, "Point", "Cr")
	d.Decode(&v)
	fmt.Println(0, "Point", "Cr", v)

	w.MoveTo(1, "Point", "G")
	d.Decode(&v)
	fmt.Println(1, "Point", "G", v)

	// Output:
	// 0 Space YCbCr
	// 0 Point Cr -10
	// 1 Point G 218
}
