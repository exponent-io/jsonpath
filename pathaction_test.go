package jsonpath

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathActionSingleMatch(t *testing.T) {

	j := []byte(`
	{
		"foo": 1,
		"bar": 2,
		"test": "Hello, world!",
		"baz": 123.1,
		"array": [
			{"foo": 1},
			{"bar": 2},
			{"baz": 3}
		],
		"subobj": {
			"foo": 1,
			"subarray": [1,2,3],
			"subsubobj": {
				"bar": 2,
				"baz": 3,
				"array": ["hello", "world"]
			}
		},
		"bool": true
	}`)

	decodeCount := 0
	decode := func(d *Decoder) interface{} {
		decodeCount++
		var v interface{}
		err := d.Decode(&v)
		assert.NoError(t, err)
		return v
	}

	dc := NewDecoder(bytes.NewBuffer(j))
	actions := &PathActions{}

	actions.Action(func(d *Decoder) {
		assert.Equal(t, float64(2), decode(d))
	}, "array", 1, "bar")

	actions.Action(func(d *Decoder) {
		assert.Equal(t, "Hello, world!", decode(d))
	}, "test")

	actions.Action(func(d *Decoder) {
		assert.Equal(t, []interface{}{float64(1), float64(2), float64(3)}, decode(d))
	}, "subobj", "subarray")

	actions.Action(func(d *Decoder) {
		assert.Equal(t, float64(1), decode(d))
	}, "foo")

	actions.Action(func(d *Decoder) {
		assert.Equal(t, float64(2), decode(d))
	}, "bar")

	dc.Scan(actions)

	assert.Equal(t, 5, decodeCount)
}

func TestPathActionAnyIndex(t *testing.T) {

	j := []byte(`
	{
		"foo": 1,
		"bar": 2,
		"test": "Hello, world!",
		"baz": 123.1,
		"array": [
			{"num": 1},
			{"num": 2},
			{"num": 3}
		],
		"subobj": {
			"foo": 1,
			"subarray": [1,2,3],
			"subsubobj": {
				"bar": 2,
				"baz": 3,
				"array": ["hello", "world"]
			}
		},
		"bool": true
	}`)

	dc := NewDecoder(bytes.NewBuffer(j))
	actions := &PathActions{}

	numbers := []int{}
	actions.Action(func(d *Decoder) {
		var v int
		err := d.Decode(&v)
		require.NoError(t, err)
		numbers = append(numbers, v)
	}, "array", AnyIndex, "num")

	numbers2 := []int{}
	actions.Action(func(d *Decoder) {
		var v int
		err := d.Decode(&v)
		require.NoError(t, err)
		numbers2 = append(numbers2, v)
	}, "subobj", "subarray", AnyIndex)

	strings := []string{}
	actions.Action(func(d *Decoder) {
		var v string
		err := d.Decode(&v)
		require.NoError(t, err)
		strings = append(strings, v)
	}, "subobj", "subsubobj", "array", AnyIndex)

	dc.Scan(actions)

	assert.Equal(t, []int{1, 2, 3}, numbers)
	assert.Equal(t, []int{1, 2, 3}, numbers2)
	assert.Equal(t, []string{"hello", "world"}, strings)
}

func TestPathActionJsonStream(t *testing.T) {

	j := []byte(`
	{
    "make": "Porsche",
		"model": "356 Coupé",
    "years": { "from": 1948, "to": 1965}
  }
  {
    "years": { "from": 1964, "to": 1969},
    "make": "Ford",
    "model": "GT40"
  }
  {
    "make": "Ferrari",
    "model": "308 GTB",
    "years": { "to": 1985, "from": 1975}
  }
  `)

	dc := NewDecoder(bytes.NewBuffer(j))

	var from, to []int
	actions := &PathActions{}
	actions.Action(func(d *Decoder) {
		var v int
		err := d.Decode(&v)
		require.NoError(t, err)
		from = append(from, v)
	}, "years", "from")
	actions.Action(func(d *Decoder) {
		var v int
		err := d.Decode(&v)
		require.NoError(t, err)
		to = append(to, v)
	}, "years", "to")

	var err error
	for err == nil {
		_, err = dc.Scan(actions)
		if err != io.EOF {
			require.NoError(t, err)
		}
	}

	assert.Equal(t, []int{1948, 1964, 1975}, from)
	assert.Equal(t, []int{1965, 1969, 1985}, to)
}

func TestPathActionJsonSubObjects(t *testing.T) {

	j := []byte(`
    {
      "set": "cars",
    	"data": [
        {
          "make": "Porsche",
      		"model": "356 Coupé",
          "years": { "from": 1948, "to": 1965}
        },
        {
          "years": { "from": 1964, "to": 1969},
          "make": "Ford",
          "model": "GT40"
        },
        {
          "make": "Ferrari",
          "model": "308 GTB",
          "years": { "to": 1985, "from": 1975}
        }
      ],
      "more": true
    }
  `)

	dc := NewDecoder(bytes.NewBuffer(j))

	var from, to []int
	actions := &PathActions{}
	actions.Action(func(d *Decoder) {
		var v int
		err := d.Decode(&v)
		require.NoError(t, err)
		from = append(from, v)
	}, "data", AnyIndex, "years", "from")
	actions.Action(func(d *Decoder) {
		var v int
		err := d.Decode(&v)
		require.NoError(t, err)
		to = append(to, v)
	}, "data", AnyIndex, "years", "to")

	var err error
	for err == nil {
		_, err = dc.Scan(actions)
		if err != io.EOF {
			require.NoError(t, err)
		}
	}

	assert.Equal(t, []int{1948, 1964, 1975}, from)
	assert.Equal(t, []int{1965, 1969, 1985}, to)
}

func TestPathActionSeekThenScan(t *testing.T) {

	j := []byte(`
    {
      "set": "cars",
    	"data": [
        {
          "make": "Porsche",
      		"model": "356 Coupé",
          "years": { "from": 1948, "to": 1965}
        },
        {
          "years": { "from": 1964, "to": 1969},
          "make": "Ford",
          "model": "GT40"
        },
        {
          "make": "Ferrari",
          "model": "308 GTB",
          "years": { "to": 1985, "from": 1975}
        }
      ],
      "more": true
    }
  `)

	dc := NewDecoder(bytes.NewBuffer(j))
	ok, err := dc.SeekTo("data", 0)
	require.NoError(t, err)
	require.True(t, ok)

	var from, to int
	actions := &PathActions{}
	actions.Action(func(d *Decoder) {
		err := d.Decode(&from)
		require.NoError(t, err)
	}, "years", "from")
	actions.Action(func(d *Decoder) {
		err := d.Decode(&to)
		require.NoError(t, err)
	}, "years", "to")

	outs := []string{}
	for err == nil {
		ok, err = dc.Scan(actions)
		if err != io.EOF {
			require.NoError(t, err)
		}
		if ok {
			outs = append(outs, fmt.Sprintf("%v-%v", from, to))
		}
	}

	assert.Equal(t, []string{"1948-1965", "1964-1969", "1975-1985"}, outs)
}

func TestPathActionSeekOffsetThenScan(t *testing.T) {

	j := []byte(`
    {
      "set": "cars",
    	"data": [
        {
          "make": "Porsche",
      		"model": "356 Coupé",
          "years": { "from": 1948, "to": 1965}
        },
        {
          "years": { "from": 1964, "to": 1969},
          "make": "Ford",
          "model": "GT40"
        },
        {
          "make": "Ferrari",
          "model": "308 GTB",
          "years": { "to": 1985, "from": 1975}
        }
      ],
      "more": true
    }
  `)

	dc := NewDecoder(bytes.NewBuffer(j))
	ok, err := dc.SeekTo("data", 1)
	require.NoError(t, err)
	require.True(t, ok)

	var from, to int
	actions := &PathActions{}
	actions.Action(func(d *Decoder) {
		err := d.Decode(&from)
		require.NoError(t, err)
	}, "years", "from")
	actions.Action(func(d *Decoder) {
		err := d.Decode(&to)
		require.NoError(t, err)
	}, "years", "to")

	outs := []string{}
	for err == nil {
		ok, err = dc.Scan(actions)
		if err != io.EOF {
			require.NoError(t, err)
		}
		if ok {
			outs = append(outs, fmt.Sprintf("%v-%v", from, to))
		}
	}

	assert.Equal(t, []string{"1964-1969", "1975-1985"}, outs)
}

func TestPathActionSeekThenScanThenScan(t *testing.T) {

	j := []byte(`
    {
      "set": "cars",
    	"data": [
        {
          "make": "Porsche",
      		"model": "356 Coupé",
          "years": { "from": 1948, "to": 1965}
        },
        {
          "years": { "from": 1964, "to": 1969},
          "make": "Ford",
          "model": "GT40"
        }
      ],
      "more": [
        {
          "make": "Ferrari",
          "model": "308 GTB",
          "years": { "to": 1985, "from": 1975}
        }
      ]
    }
  `)

	dc := NewDecoder(bytes.NewBuffer(j))
	ok, err := dc.SeekTo("data", 0)
	require.NoError(t, err)
	require.True(t, ok)

	var from, to int
	actions := &PathActions{}
	actions.Action(func(d *Decoder) {
		err := d.Decode(&from)
		require.NoError(t, err)
	}, "years", "from")
	actions.Action(func(d *Decoder) {
		err := d.Decode(&to)
		require.NoError(t, err)
	}, "years", "to")

	outs := []string{}
	for ok && err == nil {
		ok, err = dc.Scan(actions)
		if err != io.EOF {
			require.NoError(t, err)
		}
		if ok {
			outs = append(outs, fmt.Sprintf("%v-%v", from, to))
		}
	}

	assert.Equal(t, []string{"1948-1965", "1964-1969"}, outs)

	ok, err = dc.SeekTo("more", 0)
	require.NoError(t, err)
	require.True(t, ok)
	outs = []string{}
	for ok && err == nil {
		ok, err = dc.Scan(actions)
		if err != io.EOF {
			require.NoError(t, err)
		}
		if ok {
			outs = append(outs, fmt.Sprintf("%v-%v", from, to))
		}
	}

	assert.Equal(t, []string{"1975-1985"}, outs)
}
