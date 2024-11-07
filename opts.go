//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/opts
//

// Package opts is the helper library for Functional Option Pattern.
// It solves common challenges developers meet, while using this pattern at scale.
// Notably, it addresses:
//
// Boilerplate: Implementing the Option Pattern leads to significant boilerplate
// code. Each option requires a separate function, making the code more verbose,
// harder to maintain and unit test. The library automates declaration of options
// using generics.
//
// Mandatory Parameters: Enforcing mandatory parameters in an Option Pattern
// setup can be awkward. The library implements declarative approach to reuse
// the same instance of Option type to configure and validate, ensuring type
// safety and making the code easier to maintain.
//
// Compatibility with Dependency Options: When two or more libraries are chained, mapping options between dependencies can be tricky. This requires additional layers to translate options, adding more complexity.
// Discoverability: As options are applied through functions rather than struct fields, it can be harder for users to see all available configurations. Users may need to reference documentation to find all options.
package opts

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fogfish/golem/optics"
)

// Option is an abstract type for configuring instances of `S`.
// The library provides [ForType], [ForName] and [Opt] helpers to create concrete functional options.
// These helpers eliminate the need for boilerplate code when defining new options.
//
// Clients are encouraged to define type aliases for improved readability and ease of use:
//
//	type Option = opts.Option[Client]
type Option[S any] interface {
	apply(*S) error
	check(*S) error
}

// [ForType] is helper function to generate functional option for configuring
// attribute of type `A` at instances `S`.
//
// Clients typically use:
//
//	type Client { host string }
//
//	var WithHost = opts.ForType[Client, string]()
//
// By default, [ForType] creates a simple setter. The API also supports defining complex
// configuration logic within option functions. You can provide a config function
// that performs validations, type conversions, or creates new instances.
//
//	var WithHost = opts.ForType[Client, string](func(c *Client, opt string) error {
//		// e.g. validate input & return error
//	})
func ForType[S, A any](config ...func(*S, A) error) func(A) Option[S] {
	lens := optics.ForProduct1[S, A]()

	var f func(*S, A) error
	if len(config) == 1 {
		f = config[0]
	}

	return func(value A) Option[S] {
		return opt[S, A]{
			name:  fmt.Sprintf("%T", *new(A)),
			value: value,
			lens:  lens,
			f:     f,
		}
	}
}

// [ForName] is helper function to generate functional option for configuring
// attribute of type `A` at instances `S`.
//
// Clients typically use:
//
//	type Client { host string }
//
//	var WithHost = opts.ForName[Client, string]("host")
//
// By default, [ForName] creates a simple setter. The API also supports defining complex
// configuration logic within option functions. You can provide a config function
// that performs validations, type conversions, or creates new instances.
//
//	var WithHost = opts.ForName[Client, string]("host", func(c *Client, opt string) error {
//		// e.g. validate input & return error
//	})
func ForName[S, A any](attr string, config ...func(*S, A) error) func(A) Option[S] {
	lens := optics.ForProduct1[S, A](attr)

	var f func(*S, A) error
	if len(config) == 1 {
		f = config[0]
	}

	return func(value A) Option[S] {
		return opt[S, A]{
			name:  attr,
			value: value,
			lens:  lens,
			f:     f,
		}
	}
}

// [Opt] is a helper function for generating functional options to configure
// attributes of type `A` within instances of `S`. [Opt] and [ForName] are complementary:
// [ForName] returns a functional option instance that can be assigned to variables.
// However, one drawback is that options created this way appear in the "Variables"
// section of documentation. If clients want all functional options to appear
// under the "Option" type for clearer documentation, they may need to...
//
//	type Option = opts.Option[Client]
//
//	func WithHost(host string) Option { return opts.Opt[Client, string]("host", host) }
//
// Using [Opt] may add some verbosity but helps organize the documentation more effectively.
func Opt[S, A any](attr string, value A, config ...func(*S, A) error) Option[S] {
	lens := optics.ForProduct1[S, A](attr)

	var f func(*S, A) error
	if len(config) == 1 {
		f = config[0]
	}

	return opt[S, A]{
		name:  attr,
		value: value,
		lens:  lens,
		f:     f,
	}
}

type opt[S, A any] struct {
	name  string
	value A
	lens  optics.Lens[S, A]
	f     func(*S, A) error
}

//lint:ignore U1000 false positive
func (opt opt[S, A]) apply(s *S) error {
	opt.lens.Put(s, opt.value)

	if opt.f != nil {
		if err := opt.f(s, opt.value); err != nil {
			return err
		}
	}
	return nil
}

//lint:ignore U1000 false positive
func (opt opt[S, A]) check(s *S) error {
	a := opt.lens.Get(s)

	if reflect.ValueOf(a).IsZero() {
		return fmt.Errorf("undefined option %s, use With%s%s", opt.name, strings.ToUpper(opt.name[0:1]), opt.name[1:])
	}

	return nil
}

// Join multiple options to single one, creating defaults and presets.
func Join[S any](opts ...Option[S]) Option[S] { return options[S](opts) }

// [Use] is a helper function for generating functional options to configure
// instances of `S` with attributes of type `A`, where `A` itself is also
// configurable through `Option[T]` and factory `f`.
//
// Let's consider example when configurable type uses another configurable type.
//
//	type Client struct { http.Stack }
//
//	var WithHttp = opts.Use[Client](http.New)
func Use[S, A, T any](f func(...Option[T]) (A, error)) func(...Option[T]) Option[S] {
	lens := optics.ForProduct1[S, A]()

	return func(opts ...Option[T]) Option[S] {
		return make[S, A, T]{
			name: fmt.Sprintf("%T", *new(A)),
			opts: opts,
			lens: lens,
			f:    f,
		}
	}
}

type make[S, A, T any] struct {
	name string
	opts options[T]
	lens optics.Lens[S, A]
	f    func(...Option[T]) (A, error)
}

//lint:ignore U1000 false positive
func (opt make[S, A, T]) apply(s *S) error {
	a, err := opt.f(opt.opts)
	if err != nil {
		return err
	}

	opt.lens.Put(s, a)
	return nil
}

//lint:ignore U1000 false positive
func (opt make[S, A, T]) check(s *S) error {
	a := opt.lens.Get(s)

	if reflect.ValueOf(a).IsZero() {
		return fmt.Errorf("undefined option %s, use With%s%s", opt.name, strings.ToUpper(opt.name[0:1]), opt.name[1:])
	}

	return nil
}

// [FMap] is a helper function for generating functional options to configure
// attributes within instances of `S` using input type 'T'.
func FMap[S, T any](f func(*S, T) error) func(T) Option[S] {
	return func(value T) Option[S] {
		return fmap[S, T]{
			value: value,
			f:     f,
		}
	}
}

type fmap[S, T any] struct {
	value T
	f     func(*S, T) error
}

//lint:ignore U1000 false positive
func (opt fmap[S, T]) apply(s *S) error { return opt.f(s, opt.value) }

//lint:ignore U1000 false positive
func (opt fmap[S, T]) check(s *S) error { return nil }

// [From] is helper function for building default options
func From[S any](f func(*S) error) func() Option[S] {
	return func() Option[S] {
		return from[S]{
			f: f,
		}
	}
}

type from[S any] struct {
	f func(*S) error
}

//lint:ignore U1000 false positive
func (opt from[S]) apply(s *S) error { return opt.f(s) }

//lint:ignore U1000 false positive
func (opt from[S]) check(s *S) error { return nil }

// [Apply] sequence of options over the configuration type `S`.
//
//	func New(opt ...Option) (*Client, error) {
//		...
//		if err := opts.Apply(&c, opt); err != nil {
//			return nil, err
//		}
//	}
func Apply[S any](s *S, opts []Option[S]) error { return options[S](opts).apply(s) }

// [Required] checks that mandatory parameters are defined within instance of `S`.
func Required[S any](s *S, opts ...Option[S]) error { return options[S](opts).check(s) }

type options[S any] []Option[S]

func (opts options[S]) apply(s *S) error {
	for _, opt := range opts {
		if err := opt.apply(s); err != nil {
			return err
		}
	}
	return nil
}

func (opts options[S]) check(s *S) error {
	for _, opt := range opts {
		if err := opt.check(s); err != nil {
			return err
		}
	}
	return nil
}
