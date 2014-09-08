# TOML parser library for Golang [![Build Status](https://travis-ci.org/naoina/toml.png?branch=master)](https://travis-ci.org/naoina/toml)

[TOML](https://github.com/toml-lang/toml) parser library for [Golang](http://golang.org/).

This library is compatible with TOML version [v0.2.0](https://github.com/toml-lang/toml/blob/master/versions/toml-v0.2.0.md).

## Installation

    go get -u github.com/naoina/toml

## Usage

The following TOML save as `example.toml`.

```toml
title = "TOML Example"

[owner]
name = "Tom Preston-Werner"
organization = "GitHub"
bio = "GitHub Cofounder & CEO\nLikes tater tots and beer."
dob = 1979-05-27T07:32:00Z # First class dates? Why not?

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
        Org  string `toml:"organization"`
        Bio  string
        Dob  time.Time
    }
    Database struct {
        Server        string
        Ports         []int
        ConnectionMax uint
        Enabled       bool
    }
    Servers struct {
        Alpha Server
        Beta  Server
    }
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

## Documentation

See [Godoc](http://godoc.org/github.com/naoina/toml).

## License

MIT
