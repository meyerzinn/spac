# spac

Spac is a 2D MMO game written as a proof of concept for game development in Go and as an intellectual challenge.

## Getting Started

These instructions will get you a copy of the server and client up and running on your local machine for development and testing purposes.

### Prerequisites

Spac depends on [faiface/pixel](https://github.com/faiface/pixel/) for graphics, which uses OpenGL to render. Because of that, OpenGL development libraries are needed for compilation. From pixel's readme:

> The OpenGL version used is **OpenGL 3.3**.
>
> - On macOS, you need Xcode or Command Line Tools for Xcode (`xcode-select --install`) for required headers and libraries.
> - On Ubuntu/Debian-like Linux distributions, you need `libgl1-mesa-dev` and `xorg-dev` packages.
> - On CentOS/Fedora-like Linux distributions, you need `libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel mesa-libGL-devel libXi-devel` packages.
> - See [here](http://www.glfw.org/docs/latest/compile.html#compile_deps) for full details.

**NOTE**: Due to a bug with Go 1.8, compiling on macOS with XCode will fail. Updating to Go 1.8.3 resolves this issue.

### Installing

Installing spac is as simple as cloning this repo into your Go path (in the appropriate location, `$GOPATH/src/github.com/20zinnm/spac`) and running:

```
go build -o spac-server github.com/20zinnm/spac/server
```

To start the server, just run the compiled binary:

```
./spac-server
```

The CLI includes help dialogues that explain parameters.

Next, to run the client, run:

```
go build -o spac-client github.com/20zinnm/spac/client
```

Again, start the client by running the binary:

```
./spac-client
```

And voilà!

### Authors

* **Meyer Zinn** - St. Mark's

### License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

### Acknowledgments

This project would not have been possible without:

* [Michal Štrba](https://github.com/faiface/), whose excellent work on pixel helped greatly in implementing the client.
* [Jake Coffman](http://www.jakecoffman.com/), whose faithful port of the chipmunk2d physics engine contributed to the pseudo-realistic physics the game enjoys, and whose tanklets project served as a source of inspiration for spac.

