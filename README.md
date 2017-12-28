# Minecraft &nbsp;[![Travis](https://travis-ci.org/PhilipBorgesen/minecraft.svg?branch=master)](https://travis-ci.org/PhilipBorgesen/minecraft/branches#) [![Coverage Status](https://coveralls.io/repos/github/PhilipBorgesen/minecraft/badge.svg)](https://coveralls.io/github/PhilipBorgesen/minecraft) [![GoDoc](https://godoc.org/github.com/PhilipBorgesen/minecraft?status.svg)](https://godoc.org/github.com/PhilipBorgesen/minecraft) [![Dependencies](https://img.shields.io/librariesio/github/PhilipBorgesen/minecraft.svg)](https://libraries.io/github/PhilipBorgesen/minecraft) [![Go Report Card](https://goreportcard.com/badge/github.com/PhilipBorgesen/minecraft)](https://goreportcard.com/report/github.com/PhilipBorgesen/minecraft)

**Minecraft** is a collection of [SemVer][SemVerRef]-versioned Minecraft
libraries written in Go. At the time of writing it contains the following
Go packages:

  - [`profile`][ProfileRef], a binding for querying the public Mojang API
    for Minecraft profiles, supporting:
    - Lookup based on either Minecraft username or ID.
    - Fetching ID, current username, skin textures and history of prior
      usernames.
  - [`versions`][VersionsRef], a small package for fetching Mojang's
    listing of Minecraft versions and working with the reported version
    information; includes release dates of both official releases and the
    latest development snapshots.

**Examples of usage** can be found on the [GoDoc reference pages][GoDocRef]
linked above.

[SemVerRef]: http://semver.org/spec/v2.0.0.html
[ProfileRef]: https://godoc.org/github.com/PhilipBorgesen/minecraft/profile
[VersionsRef]: https://godoc.org/github.com/PhilipBorgesen/minecraft/versions
[GoDocRef]: https://godoc.org/github.com/PhilipBorgesen/minecraft

## Installing

Use `go get` to download the packages to your workspace or update them:

```sh
$ go get -u github.com/PhilipBorgesen/minecraft/...
```

## Running the tests

Use `go test` to run the entire test suite incl. integration tests:

```sh
$ go test -tags=integration github.com/PhilipBorgesen/minecraft/...
```

## Known issues

The project has been used by [the author](https://github.com/PhilipBorgesen)
to learn Go. With that said, it is not perfect:

* Tests cannot be run in parallel due to dependence on global client variables.
* [`profile.ErrMaxSizeExceeded`](https://godoc.org/github.com/PhilipBorgesen/minecraft/profile#ErrMaxSizeExceeded)
  should have been named `MaxSizeExceededError` to adhere to
  [naming conventions][NamingRef]. As an underlying design issue it did not
  need to be a type, a single error value would have been sufficient.

[NamingRef]: https://talks.golang.org/2014/names.slide#14

## License

The project is licensed under the MIT License - see [LICENSE](LICENSE) file for
details.