# tunnel-serve
[![language](https://img.shields.io/badge/language-golang?logo=go)]

Tunnel any local service to a server on the internet.
You do not need to do any configuration of routers, etc. to have your service available over the internet since an outbound connection is used on the client.

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
