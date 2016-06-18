# TOML parser and encoder library for Golang [![Build Status](https://travis-ci.org/naoina/toml.png?branch=master)](https://travis-ci.org/naoina/toml)

[TOML](https://github.com/toml-lang/toml) parser and encoder library for [Golang](http://golang.org/).

This library is compatible with TOML version [v0.4.0](https://github.com/toml-lang/toml/blob/master/versions/en/toml-v0.4.0.md).

## Installation

    go get -u github.com/naoina/toml

## Usage

The following TOML save as `example.toml`.

```toml
# This is a TOML document. Boom.

title = "TOML Example"

[owner]
name = "Lance Uppercut"
dob = 1979-05-27T07:32:00-08:00 # First class dates? Why not?

[database]
server = "192.168.1.1"
ports = [ 8001, 8001, 8002 ]
connection_max = 5000
enabled = true

[servers]

  # You can indent as you please. Tabs or spaces. TOML don't care.
  [servers.alpha]
  ip = "10.0.0.1"
  dc = "eqdc10"

  [servers.beta]
  ip = "10.0.0.2"
  dc = "eqdc10"

[clients]
data = [ ["gamma", "delta"], [1, 2] ]

# Line breaks are OK when inside arrays
hosts = [
  "alpha",
  "omega"
]
```

Then above TOML will mapping to `tomlConfig` struct using `toml.Unmarshal`.

```go
package main

import (
    "io/ioutil"
    "os"
    "time"

    "github.com/naoina/toml"
)

type tomlConfig struct {
    Title string
    Owner struct {
        Name string
        Dob  time.Time
    }
    Database struct {
        Server        string
        Ports         []int
        ConnectionMax uint
        Enabled       bool
    }
    Servers map[string]Server
    Clients struct {
        Data  [][]interface{}
        Hosts []string
    }
}

type Server struct {
    IP string
    DC string
}

func main() {
    f, err := os.Open("example.toml")
    if err != nil {
        panic(err)
    }
    defer f.Close()
    buf, err := ioutil.ReadAll(f)
    if err != nil {
        panic(err)
    }
    var config tomlConfig
    if err := toml.Unmarshal(buf, &config); err != nil {
        panic(err)
    }
    // then to use the unmarshaled config...
}
```

## Mappings

A key and value of TOML will map to the corresponding field.
The fields of struct for mapping must be exported.

The rules of the mapping of key are following:

#### Exact matching

```toml
timeout_seconds = 256
```

```go
type Config struct {
	Timeout_seconds int
}
```

#### Camelcase matching

```toml
server_name = "srv1"
```

```go
type Config struct {
	ServerName string
}
```

#### Uppercase matching

```toml
ip = "10.0.0.1"
```

```go
type Config struct {
	IP string
}
```

See the following examples for the value mappings.

### String

```toml
val = "string"
```

```go
type Config struct {
	Val string
}
```

### Integer

```toml
val = 100
```

```go
type Config struct {
	Val int
}
```

All types that can be used are following:

* int8 (from `-128` to `127`)
* int16 (from `-32768` to `32767`)
* int32 (from `-2147483648` to `2147483647`)
* int64 (from `-9223372036854775808` to `9223372036854775807`)
* int (same as `int32` on 32bit environment, or `int64` on 64bit environment)
* uint8 (from `0` to `255`)
* uint16 (from `0` to `65535`)
* uint32 (from `0` to `4294967295`)
* uint64 (from `0` to `18446744073709551615`)
* uint (same as `uint` on 32bit environment, or `uint64` on 64bit environment)

### Float

```toml
val = 3.1415
```

```go
type Config struct {
	Val float32
}
```

All types that can be used are following:

* float32
* float64

### Boolean

```toml
val = true
```

```go
type Config struct {
	Val bool
}
```

### Datetime

```toml
val = 2014-09-28T21:27:39Z
```

```go
type Config struct {
	Val time.Time
}
```

### Array

```toml
val = ["a", "b", "c"]
```

```go
type Config struct {
	Val []string
}
```

Also following examples all can be mapped:

```toml
val1 = [1, 2, 3]
val2 = [["a", "b"], ["c", "d"]]
val3 = [[1, 2, 3], ["a", "b", "c"]]
val4 = [[1, 2, 3], [["a", "b"], [true, false]]]
```

```go
type Config struct {
	Val1 []int
	Val2 [][]string
	Val3 [][]interface{}
	Val4 [][]interface{}
}
```

### Table

```toml
[server]
type = "app"

  [server.development]
  ip = "10.0.0.1"

  [server.production]
  ip = "10.0.0.2"
```

```go
type Config struct {
	Server map[string]Server
}

type Server struct {
	IP string
}
```

You can also use the following struct instead of map of struct.

```go
type Config struct {
	Server struct {
		Development Server
		Production Server
	}
}

type Server struct {
	IP string
}
```

### Array of Tables

```toml
[[fruit]]
  name = "apple"

  [fruit.physical]
    color = "red"
    shape = "round"

  [[fruit.variety]]
    name = "red delicious"

  [[fruit.variety]]
    name = "granny smith"

[[fruit]]
  name = "banana"

  [[fruit.variety]]
    name = "plantain"
```

```go
type Config struct {
	Fruit []struct {
		Name string
		Physical struct {
			Color string
			Shape string
		}
		Variety []struct {
			Name string
		}
	}
}
```

### Using `toml.UnmarshalTOML` interface

```toml
duration = "10s"
```

```go
import (
    "time"
    "strings"
)

type Config struct {
	Duration Duration
}

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalTOML(data []byte) error {
        n, err := time.ParseDuration(strings.Trim(string(data), "\""))
        d.Duration = n
        return err
}
```

## API documentation

See [Godoc](http://godoc.org/github.com/naoina/toml).

## License

MIT
