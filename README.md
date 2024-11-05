<p align="center">
  <h3 align="center">⎡🅾🅿🆃🆂⎦</h3>
  <p align="center"><strong>Effortless Functional Options: Simplify Configuration, Minimize Boilerplate.</strong></p>

  <p align="center">
    <!-- Version -->
    <a href="https://github.com/fogfish/opts/releases">
      <img src="https://img.shields.io/github/v/tag/fogfish/opts?label=version" />
    </a>
    <!-- Documentation -->
    <a href="https://pkg.go.dev/github.com/fogfish/opts">
      <img src="https://pkg.go.dev/badge/github.com/fogfish/opts" />
    </a>
    <!-- Build Status  -->
    <a href="https://github.com/fogfish/opts/actions/">
      <img src="https://github.com/fogfish/opts/workflows/build/badge.svg" />
    </a>
    <!-- GitHub -->
    <a href="http://github.com/fogfish/opts">
      <img src="https://img.shields.io/github/last-commit/fogfish/opts.svg" />
    </a>
    <!-- Coverage -->
    <a href="https://coveralls.io/github/fogfish/opts?branch=main">
      <img src="https://coveralls.io/repos/github/fogfish/opts/badge.svg?branch=main" />
    </a>
    <!-- Go Card -->
    <a href="https://goreportcard.com/report/github.com/fogfish/opts">
      <img src="https://goreportcard.com/badge/github.com/fogfish/opts" />
    </a>
  </p>
</p>

--- 

Lightweight library crafted to streamline and automate Golang's Functional Option Pattern.

## Inspiration

The [Functional Option Pattern](https://sagikazarmark.hu/blog/functional-options-on-steroids/) in Go offers several compelling benefits, making it a strong choice as my go-to standard for development. By **reducing complexity**, it eliminates the need to manage configuration structures cluttered with a mix of optional and mandatory fields. Instead, it provides a clean, functional approach to handling optional parameters. Option functions **encapsulate** complex configuration logic, such as validation or type translation, which keeps the core struct simpler and less prone to errors. Additionally, the pattern’s **flexibility** supports backward-compatible evolution without requiring multiple overloaded constructors, enabling configurations that are both concise and easy to understand. Finally, options are functions, they provide **stronger type checking** and can validate values more dynamically when applied.

There is no consensus within the Go community on the Option Pattern, as it presents several common challenges. These challenges align with my own observations from using the pattern regularly. Notably, the issues can be grouped into several categories:
* **Complexity and Boilerplate**: Implementing the Option Pattern leads to significant boilerplate code. Each option requires a separate function, making the code more verbose, harder to maintain and unit test.
* **Mandatory Parameters**: Enforcing mandatory parameters in an Option Pattern setup can be awkward. Common solutions (like post-configuration validation or combining options with required constructor arguments) add complexity and may limit the pattern's elegance.
* **Compatibility with Dependency Options**: When two or more libraries are chained, mapping options between dependencies can be tricky. This requires additional layers to translate options, adding more complexity.
* **Discoverability**: As options are applied through functions rather than struct fields, it can be harder for users to see all available configurations. Users may need to reference documentation to find all options.

`opts` is a lightweight library crafted to streamline and automate the creation of functional options. By abstracting over struct fields (leveraging capabilities like those in [golem/optics](https://github.com/fogfish/golem)), it eliminates the primary issue of boilerplate code. This approach makes defining functional options nearly as straightforward as using a struct-based configuration, reducing complexity while preserving flexibility. 

## Getting Started

- [Inspiration](#inspiration)
- [Getting Started](#getting-started)
  - [Quick example](#quick-example)
  - [Functional Option Pattern](#functional-option-pattern)
  - [Defining functional options](#defining-functional-options)
  - ["Complex" configuration logic](#complex-configuration-logic)
  - [Apply configuration](#apply-configuration)
  - [Mandatory Parameters](#mandatory-parameters)
  - [Presets and defaults](#presets-and-defaults)
  - [Dependency injections](#dependency-injections)
- [How To Contribute](#how-to-contribute)
  - [commit message](#commit-message)
  - [bugs](#bugs)
- [License](#license)


The latest version of the library is available at `main` branch of this repository. All development, including new features and bug fixes, take place on the `main` branch using forking and pull requests as described in contribution guidelines. The stable version is available via Golang modules.

Use go get to retrieve the library and add it as dependency to your application.

```bash
go get -u github.com/fogfish/opts
```

### Quick example

Example below is most simplest illustration on how to eliminate boilerplate with Functional Option Pattern.

```go
package main

import (
  "fmt"

  "github.com/fogfish/opts"
)

// Configuration type
type Client struct{ host string }

// Configuration option
var WithHost = opts.ForType[Client, string]()

// Factory creates configuration instance
func New(opt ...opts.Option[Client]) (*Client, error) {
  c := Client{}

  // apply configuration options to type
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
```

### Functional Option Pattern

The purpose of the Functional Option Pattern is a transformation of a configuration type using a sequence of functions. Given an initial configuration object C₀, a set of optional parameters is represented as a sequence of functions {ƒ₁,ƒ₂,…,ƒₙ} that each map a configuration C to a modified configuration C′. In the plain Golang code, the pattern is defined as (also see the excellent post about [Functional Option Pattern](https://sagikazarmark.hu/blog/functional-options-on-steroids/).)

```go
// Configuration type
type Client struct { … }

// Category of optional parameters
type Option func(*Client) error

// Instance of optional parameter
func WithOptA() Option { … }

// Apply optional parameters on initial configuration
func New(opts ...Option) (*Client, error) { … }
```

As we conclude, the pattern is verbose. Each option requires a separate function, making the code harder to maintain. The library defines automation receipts to streamline the pattern usage.


### Defining functional options

The functional option is a type `func(A) Option[S]` that transforms `S`. The library provides generators that automatically produce these functions, eliminating the need for clients to manually implement them.

`opts.ForType[S, A]` leverages the type hint `A` to generate an instance of the functional option for `S`. This type hint allows opts to infer the configuration target, enabling type-safe options without manual specification. By automatically aligning types, `opts.ForType` simplifies configuration and minimizes potential type mismatches, making option creation both streamlined and error-resistant.

```go
type Host string

type Client struct {
  host Host
}

var WithHost = opts.ForType[Client, Host]()
```

`opt.ForName[S, A]` uses both the type hint `A` and the specified attribute name to generate a functional option instance for `S`. By combining type and attribute name, `opt.ForName` enables precise, type-safe configuration that targets specific fields within `S`. One disadvantage of `opt.ForName` is that it can be more verbose, as it requires specifying both the type and attribute name. However, this added detail enhances clarity and reduces potential errors in complex configurations.

```go
type Client struct {
  host string
}

var WithHost = opts.ForType[Client, string]("host")
```

### "Complex" configuration logic

In 99% of cases, optional parameters function as simple setters. However, there are times when you need to perform more "complex" operations—such as validation, type conversion, or other preprocessing—before setting the value. Both `opts.ForType[S, A]` and `opts.ForName[S, A]` can optionally accept a configuration function of the form `func(*S, A) error`, allowing clients to define custom logic. This function enables additional processing, such as validation or transformation, before the value is set, giving clients fine-grained control over complex configurations. Note that the config function is executed after `S` has been configured, ensuring any dependent fields or values are available for the custom logic.

```go
type Client struct {
  n0 float64
  nL float64
}

var WithN = opts.ForName("n0", func(c *Client, n float64) error {
  if c.n0 < 0.0 {
    return fmt.Errorf("invalid n0")
  }
  
  if c.nL == 0.0 {
    c.nL = c.n0 * 2
  }

  return nil
})
```

### Apply configuration

According to the Functional Option Pattern, a constructor like `New(opt ...Option[S])` is used to accept a sequence of functional options. The client is responsible for applying each option to the configuration type. To simplify this, the library provides an `opts.Apply` helper, which automatically unwraps and applies the list of options, streamlining the configuration process.

```go
func New(opt ...Option) (*Client, error) {
  c := Client{}
  if err := opts.Apply(&c, opt); err != nil {
  }
  return c, nil
}
```

### Mandatory Parameters

Some configurations require mandatory parameters, meaning the setup should fail if any of these parameters are missing. To support this, the library provides an `opts.Required` helper function, allowing clients to specify which configuration parameters are essential. By using `opts.Required`, clients can enforce the presence of critical parameters, ensuring that configurations are validated and any missing mandatory options are detected early, preventing incomplete or invalid setups.

```go
var WithHost = opts.ForType[Client, Host]()

func (c *Client) checkRequired() error {
  return opts.Required(c, WithHost(""), /* ... */)
}

func New(opt ...Option) (*Client, error) {
  // ...
  return c, c.checkRequired()
}
```

### Presets and defaults

Certain use cases are often broad enough to be supported with pre-defined options. For configurations, this might involve bundling a set of options together to create a preset tailored to a particular use case. Presets are particularly useful for enabling a service to operate seamlessly across multiple environments. `opts.Join` groups options into single unit

```go
var (
  WithTestEnv = opts.Join(WithHost("localhost"), WithPort(8080))
  WithLiveEnv = opts.Join(WithHost("example.com"), WithPort(443))
)
```

### Dependency injections

It's common for one library to rely on the functionality of another. In Go, using interfaces and dependency injection is the recommended approach for managing these dependencies. However, in certain edge cases, it can be simpler to handle initialization directly within a top-level constructor, passing configuration options to encapsulate dependencies effectively. The library has helper `opts.From` for generating functional options to configure instances of `S` with attributes of type `A`, where `A` itself is also configurable through `Option[T]` and factory `f`.

```go
type Client struct { *http.Stack }

// The param accepts http.Option and uses http.New function to config Client
var WithHttp = opts.From[Client](http.New)

c := New(WithHost("127.1"), WithHttp(http.Timeout(5*time.Seconds)))
```

## How To Contribute

The library is [MIT](LICENSE) licensed and accepts contributions via GitHub pull requests:

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Added some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request

The build and testing process requires [Go](https://golang.org).

**build** and **test** library.

```bash
git clone https://github.com/fogfish/opts
cd opts
go test ./...
```

### commit message

The commit message helps us to write a good release note, speed-up review process. The message should address two question what changed and why. The project follows the template defined by chapter [Contributing to a Project](http://git-scm.com/book/ch5-2.html) of Git book.

### bugs

If you experience any issues with the library, please let us know via [GitHub issues](https://github.com/fogfish/opts/issue). We appreciate detailed and accurate reports that help us to identity and replicate the issue. 


## License

[![See LICENSE](https://img.shields.io/github/license/fogfish/opts.svg?style=for-the-badge)](LICENSE)
