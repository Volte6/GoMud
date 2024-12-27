# GoMud

![image](feature-screenshots/splash.png)

GoMud is an in-development open source MUD (Multi-user Dungeon) game world and library. 

It ships with a default world to play in, but can be overwritten or modified to build your own world using built-in tools.

# User Support

If you have comments, questions, suggestions:

[Github Discussions](https://github.com/volte6/gomud/discussions) - Don't be shy. Your questions or requests might help others too.

[Discord Server](https://discord.gg/cjukKvQWyy) - Get more interactive help in the GoMud Discord server.

## Screenshots

Click below to see in-game screenshots of just a handful of features:

[![Feature Screenshots](feature-screenshots/screenshots-thumb.png 'Feature Screenshots')](feature-screenshots/README.md)

## ANSI Colors

Colorization is handled through extensive use of my [github.com/Volte6/ansitags](https://github.com/Volte6/ansitags) library.

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
- [Alternate Characters](https://www.youtube.com/watch?v=VERF2l70W34)

# Quick Start

A youtube playlist to getting started has been set up here:


[![Getting Started Videos](https://i.ytimg.com/vi/OOZqX01aHt8/hqdefault.jpg 'Getting Started Playlist')](https://www.youtube.com/watch?v=OOZqX01aHt8&list=PL20JEmG_bxBuaOE9oFziAhAmx1pyXhQ1p)


You can compile and run it locally with:
> `go run .`

Or you can just build the binary if you prefer:
> `go build -o GoMudServer`

> `./GoMudServer`

Or if you have docker installed:

> `docker compose up --build`

## Connecting

*TELNET* : connect to `localhost` on port `33333` with a telnet client

*WEB CLIENT*: [http://localhost/client](http://localhost/client) 

**Default Username:** _admin_

**Default Password:** _password_

## Env Vars

When running several environment variables can be set to alter behaviors of the mud:

* **CONFIG_PATH**_=/path/to/alternative/config.yaml_ - This can provide a path to a copy of the config.yaml containing only values you wish to override. This way you don't have to modify the original config.yaml
* **LOG_PATH**_=/path/to/log.txt_ - This will write all logs to a specified file. If unspecified, will write to *stderr*.
* **LOG_LEVEL**_={LOW/MEDIUM/HIGH}_ - This sets how verbose you want the logs to be. _(Note: Log files rotate every 100MB)_

## Platform specific

### Raspberry pi

Want to run GoMud on a raspberry pi? No problem! I do it all the time! It runs great on a [$15 Raspberry Pi Zero 2](https://www.raspberrypi.com/products/raspberry-pi-zero-2-w/). However, in my experience the raspberry pi struggles to compile the binary directly, 
so it is recommended that you compile the binary locally and then copy it over to the raspberry pi.

There is a convenient `make` command to compile the pi chipset provided: 

`make build_rpi` ( this will output a binary named: `go-mud-server-rpi` )

Or (window user?) just use the build comand directly: 

`env GOOS=linux GOARCH=arm GOARM=5 go build -o go-mud-server-rpi`

# Why Go?

Why not?

Go provides a lot of terrific benefits such as:

* Compatible - High degree of compatibility across platforms or CPU Architectures. Go code quite painlessly compiles for Windows, Linux, ARM, etc. with minimal to no changes to the code.
* Fast - Go is fast. From execution to builds. The current GoMud project builds on a Macbook in less than a couple of seconds.
* Opinionated - Go style and patterns are well established and provide a reliable way to dive into a project and immediately feel familiar with the style.
* Modern - Go is a relatively new/modern language without the burden of "every feature people thought would be useful in the last 30 or 40 years" added to it. 
* Upgradable -  Go's promise of maintaining backward compatibility means upgrading versions over time remains a simple and painless process (If not downright invisible).
* Statically Linked - If you have the binary, you have the working program. Externally linked dependencies (and whether you have them) are not an issue.
* No Central Registries - Go is built to naturally incorporate library includes straight from their repos (such as git). This is neato.
* Concurrent - Go has concurrency built in as a feature of the language, not a library you include.
