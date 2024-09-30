# ltdnet
### by Vince Belanger
### https://github.com/vincebel7/ltdnet


A limited-functionality network simulator that serves as a practical application of goroutines, taking advantage of lightweight multi-threading for network traffic and device-listening.
Includes fictional LAN appliances as well as simple implementations of essential TCP/IP protocols

## How to run
1. Install Go (https://golang.org/doc/install)
2. Clone this repository and enter the directory
3. Make the run script executable, and execute it (./run.sh)

## Using ltdnet
First, create a network and pick a class (A, B, or C)
Currently, each network can only have one router. It is best to start off by creating a router:
`add router router1`

Next, you will want to create a host:
`add host host1`

You can now take a look at your network by running a show command. The most comprehensive of these is:
`show network overview`

Once you have created a host, you can "plug in" the host to the router by *linking* them.
`link host host1 router1`

This will allow you to set your host's uplink to your router.

Next, you will need to set the IP configuration for your host. There are two ways to do this: Statically, or dynamically through DHCP. Both ways require you to be in device control mode.
When controlling a device, the commands are different from the root ltdnet menu. Run `help` to see all available device control commands.
To enter device control mode, run:
`control host <hostname>`

If you wish to statically set your host's IP configuration, from device control mode run:
`ipset host <hostname>`
This will prompt you to set an IP address, subnet mask, and default gateway for your host. Remember to make sure you set the default gateway to the one your router is generated with.

To run DHCP to dynamically acquire an IP configuration, from device control mode simply run `dhcp`.

As long as there are available addresses in the router's DHCP pool, your host should now have an IP configuration. To test this out, from the device control, try pinging your router:
`ping <gateway>`

I hope you find this program to be fun. Many more features are on their way, but the main focus right now is ironing out some of the remaining bugs in v0.1 and cleaning up some debugging code still in place.

For any further questions, please email vince@vincebel.tech

## Files

### client.go
Basic menu operations, data structures

### conn.go
Handles device interfaces

### listener.go
Listens for incoming traffic and handles actions accordingly

### actions.go
All "program" functions run on devices such as ping, DHCP, etc.

### router.go
Router-specific functions and structs

### switch.go
Switch-specific functions and structs

### host.go
Host-specific functions and structs

### network.go
Network-specific functions and structs

### helpers.go
Helper functions for common operations

### debug.go
Functions for debugging and testing

### display.go
Functions for drawing ASCII network diagrams and displaying network info

### run.sh
Simple run script to run *.go

### wipesaves.sh
Clears the /saves directory

### README.md
See README.md :)

## Known bugs
- Program deadlocks on failed ping
- Save files may break with newer versions (Future fix: Version files, no breaking changes in minor versions)

## Version History
What has been accomplished at each major.minor version

### v0.2.9 - September 27, 2024
DHCP improvements, router control, router pinging

### v0.2.8 - September 25, 2024
CLI improvements, host linking bugfixes

### v0.2.7 - June 15, 2021
DHCP pool sorting changes, cleaning up JSON schema

### v0.2.6 - October 9, 2020
Increase debug output

### v0.2.5 - September 30, 2020
Improved host unlinking

### v0.2.4 - October 9, 2019
Implement virtual switch for router, ping improvements

### v0.2.3 - October 7, 2019
Deleting hosts, improved diagrams

### v0.2.2 - October 6, 2019
Network diagram, initial design of switches

### v0.2.1 - October 4, 2019
Debug system, IP configuration clearing

### v0.2 - October 4, 2019
ARP, DHCP, complete network functionality, project structuring

### v0.1 - July 25, 2019
Creating, saving, and loading JSON savefiles, creating a network, adding and linking hosts, adding routers, host controlling, device listeners, basic host-to-host pinging

### v0.0 - July 17, 2019
Basic interface

