# tunnel-serve
[![language](https://img.shields.io/badge/language-golang-blue?logo=go)](https://go.dev/)


## Table of Contents
- [About](#-about)
- [Examples](#-examples)
- [Directory Structure](#directory-structure)
- [How to Build](#how-to-build)

## About
Tunnel any local service to a server on the internet.
You do not need to do any configuration of routers, etc. to have your service available over the internet since an outbound connection is used on the client.

## Examples
Say you want to run a proxy server on your local network and make it instantly available over the internet.
1. Go into client-proxy-run

2. Configure the JSON to point to the public gateway:
```json
{
    "ProxyPort": "7070",
    "ReverseTunnelAddr": "157.230.120.90:5030",
    "LocalServiceAddr": "127.0.0.1:7070"
}
```
_"ProxyPort"_: the port the local proxy server should run on.
It is an HTTP(S) proxy.

_"ReverseTunnelAddr"_: address of the reverse tunnel (e.g. our gateway).
Our gateway is at "157.230.120.90". This is the public gateway server.
Your proxy server is configured with random credentials (HTTP Basic Authentication),
so no one else can access your proxy.

_"LocalServiceAddr"_: this is the address that your client will tunnel to.
It needs to stay at _localhost_ port _7070_ since _7070_ is where your
proxy is running.

3. Install go

4. Build the client.
```shell
go get .
go build -o client-proxy-run client-proxy-run.go
```

5. Run the client.

6. In the output, the address of your proxy server on a public IP will be given.

7. Congrats, you can access your proxy server from anywhere in the world!.


## Directory Structure
- client  
Client which serves your local service to the gateway.

- server  
Server which tunnels incoming connections to your local service.

- client-run  
Code to run the client.

- server-run  
Code to run the server.

Example app which you could serve:
ssh -D 1337 -N localhost  
(local SOCKS proxy server).  

Adjust the config.json in client-run and server-run as needed.

## How to Build

1. Install go (https://go.dev/)
2. Build server:
```shell
cd server-run
go get .
go build -o server-run server-run.go
```

3. Build client:
```shell
cd client-run
go get .
go build -o client-run client-run.go
```

4. Build client with proxy:
```shell
cd client-proxy-run
go get .
go build -o client-run client-run.go
```

## Run Example
1. Configure the server:
server-run/config.json:
```json
{
    "BindAddress": "0.0.0.0",
    "ControlPort": "5030",
    "DomainName": "192.168.1.104"
}
```

_BindAddress_: IP address the server should listen on.
Use _0.0.0.0_ to listen on all available IP addresses.

_ControlPort_: Port the server should listen on. Needs to correspond 
with _ReverseTunnelAddr_ in the client config.json.

_DomainName_: Domain name (if running locally, just input your local network IP here. Or _127.0.0.1_ if running client and server on the same machine.) This domain name will be given to the client as information on how to reach the gateway.

2. Run the server:
```shell
cd server-run
./server-run
```

3. Configure the client:
client-run/config.json:
```json
{
    "ReverseTunnelAddr": "157.230.120.90:5030",
    "LocalServiceAddr": "127.0.0.1:7070"
}
```

_ReverseTunnelAddr_: Address of the server we just started running.
_LocalServiceAddr_: Address of the service which should be made available via the server.

4. Run the client:
```shell
cd client-run
./client-run
```