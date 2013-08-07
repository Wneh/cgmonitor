#CGMonitor
Web interface to monitor several cgminer miners

##Installation

###Compiled version
Go to [releases](https://github.com/Wneh/cgmonitor/releases)

###Compile from source
Start by installing [Go](http://golang.org/doc/install) if you don't already got it. Make sure that your [work environment](http://golang.org/doc/code.html) properly done. You will also need to install Git.

Start with downloading cgmonitor source:

    $ go get github.com/Wneh/cgmonitor

Change to cgmonitor folder:

    $ cd $GOPATH/github.com/Wneh/cgmonitor

And finally build it:

    $ go build

Before you start cgmonitor you must add the miners to cgmonitor.conf([example config with two miners](https://github.com/Wneh/cgmonitor/blob/master/exampleConfig.conf)). You also need to allow the computer that will run cgmonitor to access the cgminer api.

If you use arguments:

    --api-allow W:<ip to computer that host cgmonitor>

or add these lines to the your cgminer config file:

    "api-allow" : "W:<ip to computer that host cgmonitor>",
    "api-listen" : true

Now start cgmonitor:

    $ ./cgmonitor

Now start your browser and navigate to `http://<ip address>:8080`

##Dependencies
Following dependencies are needed:

Gorilla Mux

    $ go get github.com/gorilla/mux


##License
MIT - see license file
