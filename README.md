#ltdnet
###by Vince Belanger

A limited-functionality network simulator that serves as a practical application of goroutines, taking advantage of lightweight multi-threading for network traffic.
Will include fictional LAN appliances as well as simple applications of essential TCP/IP protocols

##Files

###client.go
Basic menu operations, data structures

###conn.go
Handles device interfaces

###listener.go
Listens for incoming traffic and handles actions accordingly

##actions.go
All "program" functions run on devices such as ping, DHCP, etc.

##helpers.go
Helper functions for common operations

##run.sh
Simple run script to run *.go

##wipesaves.sh
Clears the /saves directory

##README.md
See README.md :)

##Version History
What has been accomplished at each major.minor version

###v0.0 - July 17, 2019
Basic interface

###v0.1
Creating, saving, and loading JSON savefiles, creating a network, adding and linking hosts, adding routers, host controlling, device listeners, basic host-to-host pinging
