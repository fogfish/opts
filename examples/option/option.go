//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/opts
//

package main

import (
	"fmt"

	"github.com/fogfish/opts"
)

// Configuration type
type Client struct{ host string }

// Configuration option
var WithHost = opts.ForName[Client, string]("host")

// Factory creates configuration instance
func New(opt ...opts.Option[Client]) (*Client, error) {
	c := Client{}

	// apply configuration options to instance
	if err := opts.Apply(&c, opt); err != nil {
		return nil, err
	}
	return &c, nil
}

func main() {
	c, err := New(WithHost("example.com"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("==> %+v\n", c)
}
