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
