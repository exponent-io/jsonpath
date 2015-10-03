[![GoDoc](https://godoc.org/github.com/exponent-io/json-seek?status.svg)](https://godoc.org/github.com/exponent-io/json-seek)
[![Build Status](https://travis-ci.org/exponent-io/json-seek.svg?branch=master)](https://travis-ci.org/exponent-io/json-seek)

# json-seek

Extends the Go runtime's json.Decoder enabling navigation of a stream of json tokens. This package includes a SeekingDecoder which wraps a json.Decoder and extends it to include the SeekTo() method. You should be able to use the SeekingDecoder in any place where a json.Decoder would have been used.

## Installation

    go get -u github.com/exponent-io/json-seek

## Usage

    import "github.com/exponent-io/json-seek"

    var j = []byte(`[
      {"Space": "YCbCr", "Point": {"Y": 255, "Cb": 0, "Cr": -10}},
      {"Space": "RGB",   "Point": {"R": 98, "G": 218, "B": 255}}
    ]`)

    w := json_seek.NewSeekingDecoder(bytes.NewReader(j))
    var v interface{}

    w.SeekTo(1, "Point", "G")
    w.Decode(&v) // v is 218
