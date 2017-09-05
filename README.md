GoTalk
======

More than anything this is a project to get learning the fundamentals of GoLang. I was trying to think of something to make and came across the suggestion of a simple chat application. Having never made anything like it before I am basing it quite heavily off a couple of sources I found online, most notably for the server [this project](https://github.com/coolspeed/century/blob/master/century.go).

There are now two programs to the client, this is essentially becuase the messages sent from the client, if it were a single file, would be returned to the client, and this was annoying me. So now there is GoTalk which is for sending messages tagged with a username, GoListen for listening to the conversation, and GoServe for hosting the server.

## Compile

There is currently a server and two required client programs.

```bash
mkdir logs
go build GoServe.go
./GoServe

go build GoTalk.go
./GoTalk [-ip hostIP] [-p port] [-n Username]

go build GoListen.go
./GoListen [-ip hostIP] [-p port]
```

## Usage

**Server - GoServe**
- Provide the server application with a port and IP address via command line arguments
- Once the server has started, the logs should begin to appear in ./logs/; logs information is also printed to the command line
- Logs messages received and user activity in logs/ via `glog`

```
Usage of GoServe:
	-ip string
		Host IP (default "127.0.0.1")
	-p string
		Host port (default "6666")
```

**Client - GoTalk**
- Provide GoTalk with an IP and a port via command line arguments
- GoTalk also requires a username to let others know who sent what
- Send messages and view them in a GoListen instance

```
Usage of GoTalk:
	-n string
		Username (default "BodneyC")
	-ip string
		Host IP (default "127.0.0.1")
	-p string
		Host port (default "6666")
```
**Client - GoListen**
- Provide GoTalk with an IP and a port via command line arguments
- Receive messages sent via the server (telnet/GoTalk etc)

```
Usage of GoListen:
	-ip string
		Host IP (default "127.0.0.1")
	-p string
		Host port (default "6666")
```
