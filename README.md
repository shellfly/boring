# Boring

> "What I cannot create, I do not understand."
>
> -- Richard Feynman

A toy project to help myself understand [socks5 protocol](https://en.wikipedia.org/wiki/SOCKS) and [TCP tunneling](https://en.wikipedia.org/wiki/Tunneling_protocol). The name is inspired by [The Boring Company](https://www.boringcompany.com/)

Both Client and Server are working as a TCP proxy. The Client implements socks5 protocol to accept socks5 incoming proxy request and then send the request to Server, the Server access target host and copy data back to client. The communication between Client and Server is encrypted by specified method and key.

`Application` <- socks5 -> `Client` <= encrypted data => `Server` <- tcp -> `Internet`

## Usage

### Build

``` bash
make
```

### Run server

A tcp server handling encryption connection

``` bash
# running on Linux
./server-linux -method aes -key {key}
```

### Run client

A socks5 server handling incoming proxy request, encryption data and send to server.

``` bash
# running on Mac
./client-darwin -method aes -key {key}
```
