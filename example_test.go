package jsonpath

import (
	"bytes"
	"fmt"
	"io"
)

func ExampleDecoder_SeekTo() {

	var j = []byte(`[
		{"Space": "YCbCr", "Point": {"Y": 255, "Cb": 0, "Cr": -10}},
		{"Space": "RGB",   "Point": {"R": 98, "G": 218, "B": 255}}
	]`)

	w := NewDecoder(bytes.NewReader(j))
	var v interface{}

	w.SeekTo(0, "Space")
	w.Decode(&v)
	fmt.Printf("%v => %v\n", w.Path(), v)

	w.SeekTo(0, "Point", "Cr")
	w.Decode(&v)
	fmt.Printf("%v => %v\n", w.Path(), v)

	w.SeekTo(1, "Point", "G")
	w.Decode(&v)
	fmt.Printf("%v => %v\n", w.Path(), v)

	// seek to the end of the object
	w.SeekTo()
	fmt.Printf("%v\n", w.Path())

	// Output:
	// [0 Space] => YCbCr
	// [0 Point Cr] => -10
	// [1 Point G] => 218
	// []
}

func ExampleDecoder_Scan() {

	var j = []byte(`{"colors":[
		{"Space": "YCbCr", "Point": {"Y": 255, "Cb": 0, "Cr": -10, "A": 58}},
		{"Space": "RGB",   "Point": {"R": 98, "G": 218, "B": 255, "A": 231}}
	]}`)

	var actions PathActions

	// Extract the value at Point.A
	actions.Add(func(d *Decoder) error {
		var alpha int
		err := d.Decode(&alpha)
		fmt.Printf("Alpha: %v\n", alpha)
		return err
	}, "Point", "A")

	w := NewDecoder(bytes.NewReader(j))
	w.SeekTo("colors", 0)

	var ok = true
	var err error
	for ok {
		ok, err = w.Scan(&actions)
		if err != nil && err != io.EOF {
			panic(err)
		}
	}

	// Output:
	// Alpha: 58
	// Alpha: 231
}

func ExampleDecoder_Scan_anyIndex() {

	var j = []byte(`{"colors":[
		{"Space": "YCbCr", "Point": {"Y": 255, "Cb": 0, "Cr": -10, "A": 58}},
		{"Space": "RGB",   "Point": {"R": 98, "G": 218, "B": 255, "A": 231}}
	]}`)

	var actions PathActions

	// Extract the value at Point.A
	actions.Add(func(d *Decoder) error {
		var cr int
		err := d.Decode(&cr)
		fmt.Printf("Chrominance (Cr): %v\n", cr)
		return err
	}, "colors", AnyIndex, "Point", "Cr")

	actions.Add(func(d *Decoder) error {
		var r int
		err := d.Decode(&r)
		fmt.Printf("Red: %v\n", r)
		return err
	}, "colors", AnyIndex, "Point", "R")

	w := NewDecoder(bytes.NewReader(j))

	var ok = true
	var err error
	for ok {
		ok, err = w.Scan(&actions)
		if err != nil && err != io.EOF {
			panic(err)
		}
	}

	// Output:
	// Chrominance (Cr): -10
	// Red: 98
}
