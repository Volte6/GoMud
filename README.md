# GoMud

This is an early version of GoMud, having only been in development a couple of months.

It has been refactored on the fly, which is why some aspects might seem less than ideal.

The network layer still needs to be cleaned up, since it start with very different assumptions than where things ended up, but it works fine as it is for now.

Screenshots for some of the features can be found [here](https://imgur.com/a/90y6OGS).

Colorization is handled through extensive use of my [ansitags](https://github.com/Volte6/ansitags) library.

Can be run locally as a standard go program or via docker container. The default port is 33333.

There is not yet anything for the web service side of things, nor does the web client work yet.

Certain admin in-game commands can be destructive. For example, the `build` command is notoriously finicky if you don't understand what you are doing. Although there is some documentation, it doesn't mean stuff won't get missed, plus it's possible to accidentally mess up typing something and then can be tricky to fix if you don't first understand the underlying mechanisms.

Plans later when a network-layer overhaul takes place is to enable exclusive graphics modes (full screen takeover without scrollback) for some stuff, as well as a multi-prompt interaction, where all input/output is isolated until a series of prompts are completed (or aborted). This will allow for better tools which won't require all information to be present in one command, or split into multiple individual commands. It will also allow for confirmation dialogs etc.

# There is one default user created:

*Username:* _admin_

*Password:* _password_

## Running locally:

You can compile and run it locally with:
> `go run .`

## Makefile usage

docker run -u root --name tmp -it alpine:3.14;docker rm tmp
exec -it <container name> /bin/bash

_________________

### **Go Specific Makefile targets**

_These require Go to be installed locally_

Run go vendor/tidy/verify:
> `make mod`

Run in a container (port 33333):
> `make run`

Connect to running container via a container client:
> `make client`

_________________

### **To Restart Docker Daemon**

_From powershell w/ admin priv:_

> `restart-service *docker*`
_________________

### **Dockerfile specific Makefile targets**

_These require Docker to be installed locally_

Build/Run in Docker container:

_Will run on port 8080 in the container and publicly exposes port 8080 ( per docker-compose.yml ):_

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

