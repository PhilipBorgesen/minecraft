# Overview
This is a collection of [SemVer](http://semver.org/spec/v2.0.0.html)-versioned Minecraft libraries written in Go.
At the time of writing it contains the following Go packages:

  - `profile`, a binding for querying the public Mojang API for Minecraft profiles, supporting:
    - Lookup based on either Minecraft username or ID.
    - Fetching ID, current username, skin textures and history of prior usernames.
    - Caching.
  - `versions`, a small package for fetching Mojang's listing of Minecraft versions
    and working with the reported version information. Most important this package
    reports the latest snapshot and release versions of Minecraft.

Documentation can be found at https://godoc.org/github.com/PhilipBorgesen/minecraft.

TODOs & Ideas
-------------
  - `profile/cache`, a package with a simple in-memory cache implementation of `profile.Cache`.
    - Once written, add reference in `profile` to package for an implementation example.