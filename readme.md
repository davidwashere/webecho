# Web Echo
Simple web-server that will echo the request it recieves to stdout and the http response

It includes a web gui for viewing historic requests (via the admin port)

## Usage
```
$ ./webecho
2024/03/27 00:37:48   Web Server listening on port 8080
2024/03/27 00:37:48 Admin Server listening on port 8081
```

To customize the ports:

```
$ ./webecho
Usage of webecho
  -adminport int
        admin server listen port (default 8081)
  -webport int
        web server listen port (default 8080)

$ ./webecho --webport 9999 --adminport 10000
2024/03/27 00:40:35   Web Server listening on port 9999
2024/03/27 00:40:35 Admin Server listening on port 10000
```

Making a simple request
```
$ curl -s localhost:5080?hello=world
GET /?hello=world HTTP/1.1
Host: localhost:5080
User-Agent: curl/7.65.3
Accept: */*
```

Also shown in the log
```
2020/02/17 00:48:33 WEB: [127.0.0.1:64939] /:

GET /?hello=world HTTP/1.1
Host: localhost:5080
User-Agent: curl/7.65.3
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