//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/opts
//

package opts_test

import (
	"testing"

	"github.com/fogfish/it/v2"
	"github.com/fogfish/opts"
)

type Option = opts.Option[Client]

type Host string

type Client struct {
	host Host
	addr string
}

func New(opt ...Option) (*Client, error) {
	c := Client{}
	if err := opts.Apply(&c, opt); err != nil {
		return nil, err
	}
	return &c, nil
}

const kHost = "example.com"
const kAddr = "127.0.0.1"

func TestForType(t *testing.T) {
	t.Run("Type", func(t *testing.T) {
		withHost := opts.ForType[Client, Host]()
		c, err := New(withHost(kHost))

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(c.host, kHost),
		)
	})

	t.Run("WithConfig", func(t *testing.T) {
		withHost := opts.ForType(
			func(c *Client, h Host) error { return nil },
		)
		c, err := New(withHost(kHost))

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(c.host, kHost),
		)
	})

}

func TestForName(t *testing.T) {
	t.Run("Name", func(t *testing.T) {
		withHost := opts.ForName[Client, string]("addr")
		c, err := New(withHost(kAddr))

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(c.addr, kAddr),
		)
	})

	t.Run("WithConfig", func(t *testing.T) {
		withHost := opts.ForName("addr",
			func(c *Client, h string) error { return nil },
		)
		c, err := New(withHost(kAddr))

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(c.addr, kAddr),
		)
	})
}

func TestOpt(t *testing.T) {
	t.Run("Name", func(t *testing.T) {
		withHost := func(x string) Option {
			return opts.Opt[Client]("addr", x)
		}
		c, err := New(withHost(kAddr))

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(c.addr, kAddr),
		)
	})

	t.Run("WithConfig", func(t *testing.T) {
		withHost := func(x string) Option {
			return opts.Opt("addr", x,
				func(c *Client, h string) error { return nil },
			)
		}
		c, err := New(withHost(kAddr))

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(c.addr, kAddr),
		)
	})
}

func TestJoin(t *testing.T) {
	withHost := opts.ForType[Client, Host]()
	preset := withHost(kHost)

	c, err := New(preset)

	it.Then(t).Should(
		it.Nil(err),
		it.Equal(c.host, kHost),
	)
}

type T struct {
	client *Client
}

func NewT(opt ...opts.Option[T]) (*T, error) {
	t := T{}
	if err := opts.Apply(&t, opt); err != nil {
		return nil, err
	}
	return &t, nil
}

func TestFrom(t *testing.T) {
	withHost := opts.ForType[Client, Host]()
	withClient := opts.From[T](New)

	c, err := NewT(withClient(withHost(kHost)))

	it.Then(t).Should(
		it.Nil(err),
		it.Equal(c.client.host, kHost),
	)
}

func TestRequired(t *testing.T) {
	withHost := opts.ForType[Client, Host]()
	withAddr := opts.ForName[Client, string]("addr")

	c, err := New(withHost(kHost))
	it.Then(t).Should(it.Nil(err))

	err = opts.Required(c, withHost(""))
	it.Then(t).Should(it.Nil(err))

	err = opts.Required(c, withAddr(""))
	it.Then(t).ShouldNot(it.Nil(err))
}
