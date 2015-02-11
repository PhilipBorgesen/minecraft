# Overview
This is a collection of Minecraft libraries written in Go.
At the time of writing, it only contains a single Go package:

  - `profile`, a binding for querying the public Mojang API for Minecraft profiles, including support for:
    - Lookup based on either Minecraft username or ID.
    - Fetching ID, current username, skin textures and history of prior usernames.

Documentation can be found on https://godoc.org/github.com/PhilipRasmussen/minecraft.

TODOs & Ideas
-------------
  - `profile`:
    - Testing:
      - Test code on profiles having a longer username history than one.
      - Write automatic tests for handling profiles using the slim player model.
      - Write automatic tests for handling profiles with no custom skin.
    - Add caching hooks + usage example

License
-------
Included libraries are open source dual licensed under MIT/GPL.

  - GPL license: http://www.gnu.org/licenses/gpl.html
  - MIT license: http://www.opensource.org/licenses/mit-license.php

Copyright (c) 2015 Philip Rasmussen