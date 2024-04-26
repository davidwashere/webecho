# Web Echo
Simple web-server that will echo the request it receives to stdout and the http response

It includes a web 'admin' gui for viewing historic requests

## Install
Download pre-build binaries from [releases](https://github.com/davidwashere/webecho/releases/latest), also available as a container image on [Docker Hub](https://hub.docker.com/r/davidwashere/webecho)

## Usage
```
$ ./webecho
2024/03/27 00:37:48   Web Server listening on port 8080
```

```
$ ./webecho -h
Usage of webecho
  -adminport string
        admin server listen port, when set enables admin server
  -port int
        web server listen port (default 8080)
```
```
$ ./webecho --adminport 10000
2024/03/27 00:40:35   Web Server listening on port 8080
2024/03/27 00:40:35 Admin Server listening on port 10000
```

Making a simple request
```
$ curl -s localhost:8080?hello=world
GET /?hello=world HTTP/1.1
Host: localhost:8080
User-Agent: curl/8.2.1
Accept: */*
```

Also shown in the log
```
2020/02/17 00:48:33 WEB: [127.0.0.1:64939] /:

GET /?hello=world HTTP/1.1
Host: localhost:8080
User-Agent: curl/8.2.1
Accept: */*
```

As well is in the web-gui

![](adminui.png)

## Building
To test locally execute:

```
go run .
```

To build:

```
go build .
```

Also avail on [dockerhub](https://hub.docker.com/r/davidwashere/webecho)