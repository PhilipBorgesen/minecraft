# Overview
This is a collection of Minecraft libraries written in Go.
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
  - `profile`:
    - Testing:
      - Test code on profiles having a longer username history than one.
      - Write automatic tests for handling profiles using the slim player model.
      - Write automatic tests for handling profiles with no custom skin.
  - `profile/cache`, a package with a simple in-memory cache implementation of `profile.Cache`.
    - Once written, add reference in `profile` to package for an implementation example.

License
-------
Included libraries are open source dual licensed under MIT/GPL.

  - GPL license: http://www.gnu.org/licenses/gpl.html
  - MIT license: http://www.opensource.org/licenses/mit-license.php

Copyright (c) 2016 Philip Børgesen
