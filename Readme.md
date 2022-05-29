# httpserver_redirect

httpserver_redirect is a Go program that provides you with redirect concept.

## Prerequisites

Install go  before run 

and

put file (index.html) on C:\inetpub\wwwroot\new 



## Usage

```
##Server
...>>go run httpserver3.go new
Server Running...
Listening on localhost:9980
Waiting for client...
A new client connected
Read a new line from connection: 'GET / HTTP/1.1'
line GET / HTTP/1.1
Got new 'GET' request for /

127.0.0.1:9980
Sent response with status code: 302


##Client (browser)
http://127.0.0.1:9980/
Address does not exist, Click to redirect Index.html

when user click on this address , first check existance of the file and then redirect to index.html



```

