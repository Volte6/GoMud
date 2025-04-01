# Getting Running on Various Platforms

These guides assume you want to build from the source. You can also download the latest package and run the appropriate pre-compiled binary for your platform from the [releases section](https://github.com/Volte6/GoMud/releases) of the repo.

- [Raspberry PI Zero 2W](RASPBERRY-PI.md)
- [Running via Docker](DOCKER.md)
- [Setting Up an EC2 Instance](EC2.md)


# Quick Start

You can download the latest release from the [releases page](https://github.com/Volte6/GoMud/releases), unzip it and run the binary to get started, or if you prefer to build it yourself, follow the instructions below.

A youtube playlist to getting started has been set up here:

[![Getting Started Videos](https://i.ytimg.com/vi/OOZqX01aHt8/hqdefault.jpg "Getting Started Playlist")](https://www.youtube.com/watch?v=OOZqX01aHt8&list=PL20JEmG_bxBuaOE9oFziAhAmx1pyXhQ1p)

You can compile and run it locally with:

> `go run .`

Or you can just build the binary if you prefer:

> `go build -o GoMudServer`

> `./GoMudServer`

Or if you have docker installed:

> `docker compose up --build`