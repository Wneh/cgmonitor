#CGMonitor
Web interface to monitor several cgminer instances

##Installation

###Compiled version
Go to [releases](https://github.com/Wneh/cgmonitor/releases)

###Compile from source
Start by installing [Go](http://golang.org/doc/install) if you don't already got it.

Make sure that your [work environment](http://golang.org/doc/code.html) is correct

Start with downloading cgmonitor source:

    $ go get github.com/Wneh/cgmonitor

Go to cgmonitor folder:

    $ cd $GOPATH/github.com/Wneh/cgmonitor

And finally build it:

    $ go build

Before you start cgmonitor you must add the miners to cgmonitor.conf

When your done with the config file start cgmonitor:

    $ ./cgmonitor

Now start your browser and navigate to http://<ip address>:8080

##Dependencies
Following dependencies are needed:

Gorilla Mux
```
$ go get github.com/gorilla/mux
```

##License
MIT - see license file
