# GoMud

![image](images/login.png)

This is an early version of GoMud, having only been in development a couple of months.

Screenshots for some of the features can be found [here](webclient/images/).

Colorization is handled through extensive use of my [github.com/Volte6/ansitags](https://github.com/Volte6/ansitags) library.

Can be run locally as a standard go program or via docker container. The default port is `33333`.

There is some stubbed out folders/files/code bits for a web service and web client, but nothing substantial or even moderately functional yet. Later this should use websockets to connect, and be able to server game-aware pages up.

NOTE: Certain admin in-game commands can be destructive. For example, the `build` command is notoriously finicky if you don't understand what you are doing. Although there is some documentation, it doesn't mean stuff won't get missed, plus it's possible to accidentally mess up typing something and then can be tricky to fix if you don't first understand the underlying mechanisms. Now that there is a user prompt system working this can probably be improved considerably in the near future, and building or modifying a room can be a series of prompts.

The network layer will eventually be overhauled and possibly include support for the `alternative screen buffer` mode at some point.

## Scripting

Information on scripting in GoMud can be found in the [scripting README](scripting/README.md).

## Small Feature Demos

- [Auto-complete input](https://youtu.be/7sG-FFHdhtI)
- [In-game maps](https://youtu.be/navCCH-mz_8)
- [Quests / Quest Progress](https://youtu.be/3zIClk3ewTU)
- [Lockpicking](https://youtu.be/-zgw99oI0XY)
- [Hired Mercs](https://youtu.be/semi97yokZE)
- [TinyMap](https://www.youtube.com/watch?v=VLNF5oM4pWw) (okay not much of a "feature")
- [256 Color/xterm](https://www.youtube.com/watch?v=gGSrLwdVZZQ)
- [Customizable Prompts](https://www.youtube.com/watch?v=MFkmjSTL0Ds)
- [Mob/NPC Scripting](https://www.youtube.com/watch?v=li2k1N4p74o)
- [Room Scripting](https://www.youtube.com/watch?v=n1qNUjhyOqg)
- [Kill Stats](https://www.youtube.com/watch?v=4aXs8JNj5Cc)
- [Searchable Inventory](https://www.youtube.com/watch?v=iDUbdeR2BUg)
- [Day/Night Cycles](https://www.youtube.com/watch?v=CiEbOp244cw)
- [Web Socket "Virtual Terminal"](https://www.youtube.com/watch?v=L-qtybXO4aw)

# Quick Start

A youtube playlist to getting started has been set up here:

[![Getting Started Videos](https://i.ytimg.com/vi/OOZqX01aHt8/hqdefault.jpg)](https://www.youtube.com/watch?v=OOZqX01aHt8&list=PL20JEmG_bxBuaOE9oFziAhAmx1pyXhQ1p)

You can compile and run it locally with:
> `go run .`

Or you can just build the binary if you prefer:
> `go build -o GoMudServer`

> `./GoMudServer`

Or if you have docker installed:
> `docker-compose -f provisioning/docker-compose.yml up --build --remove-orphans server`

From there you should see some logging, and once ready, connect to `localhost` on port `33333` with a telnet client and use the default admin login:

**Username:** _admin_

**Password:** _password_

## Makefile usage

There are a number of make targets that might be useful for building/running the MUD.

You can type `make help` to see a couple make targets worth knowing about.

_________________

### **Go Specific Makefile targets**

_These require Go to be installed locally_

Run go vendor/tidy/verify:
> `make mod`

Run in a container (port 33333):
> `make run`

Connect to running container via a container client:
> `make client`

Build the `go-mud-server` binary:
> `make build`

_________________

### **To Restart Docker Daemon**

_From powershell w/ admin priv:_

> `restart-service *docker*`
_________________

### **Dockerfile specific Makefile targets**

_These require Docker to be installed locally_

Build/Run in Docker container:

_Will run on port `33333` in the container and publicly exposes port `33333` ( per [provisioning/dockerdocker-compose.yml](dockerdocker-compose.yml) ):_

>  `make run`



Get exposed port of running container:

>  `make port`


_________________

## Connecting to server

_Connection can be made with any terminal program (telnet, nc, etc)_
>  `telnet localhost 33333`

_________________
_NOTE:_ Windows default telnet client is no longer compatible with typical ANSI Escape codes.
_________________


### **Connect using custom client** 

_This will build a lightweight linux container and use it to telnet to the server. This is useful especially on windows where ANSI color escape sequences are borked._
> `make client`

_Or using docker (localhost connection)_
>  `docker run --rm -u root -it busybox:latest telnet host.docker.internal 33333`

_Or using docker (external connection)_
>  `docker run --rm -u root -it busybox:latest telnet {hostname/ip} 33333`
>
[Some Notes](notes.md)

