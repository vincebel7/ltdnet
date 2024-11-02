# ltdnet
### by Vince Belanger
### https://github.com/vincebel7/ltdnet


A little CLI-based network simulator for learning networking concepts, utilizing lightweight goroutines and channels for network traffic.
Includes fictional network appliances and simple implementations of essential TCP/IP protocols.

## How to run (binary)
Download the Windows or Linux executable binary from the [Releases](https://github.com/vincebel7/ltdnet/releases) page

## How to run (from source)
1. Install Go (https://golang.org/doc/install)
2. Clone this repository and enter the directory
3. Make the run script executable, and execute it (./run.sh)

## ltdnet Tutorial
First, create a network and pick a network size (/24 is recommended to start)
Currently, each network can only have one router. It is best to start off by creating the router:
`add router router1`

Next, create a host:
`add host host1`

You can now look at your network by running a show command. The most comprehensive of these is:
`show network overview`

Once you have created a host, you can "plug in" the host to the router by *linking* them.
`link host host1 router1`

Next, set the IP configuration for your host. There are two ways to do this: Statically, or dynamically through DHCP. Both ways require you to be in device control mode.
When controlling a device, the commands are different from the main ltdnet menu. Run `help` to see all available device control commands.
To enter device control mode for host1, run:
`control host host1`

If you wish to statically set your host's IP configuration, from device control mode run:
`ip set <address>`
The program will pick a subnet mask and default gateway, then ask you to confirm the new IP configuration.

To run DHCP to dynamically acquire an IP configuration, from device control mode simply run `dhcp`. As long as there are available addresses in the router's DHCP pool, your host should now have an IP configuration.

To test connectivity, try pinging your router (likely 192.168.0.1) from the device control:
`ping <gateway>`

What now? You can start by reading the user manual from within the program with the `manual` command, or you can try to complete all the Achievements (`achievements` command).

I hope you find this program to be fun. Many more features are on their way, but the main focus right now is ironing out some of the remaining bugs in v0.x.

For any further questions, please email vincebel@protonmail.com


## Known bugs (highest to lowest impact)
- Switches can't yet connect to routers, or other switches

- Display functions sometimes cause program to crash if a host has been deleted

- Save files may break with newer versions (Future fix: No breaking changes in minor versions after v1.0)

## Scripts

### run.sh
Simple run script to run *.go from source

### wipesaves.sh
Clears the /saves directory

### build.sh
Builds executables for Linux and Windows

## Version History
What has been accomplished at each version

### v0.4.0 - November 2, 2024
Layer 1 rework (interfaces), host-level routing, static hostnames, bypasses for self-bound traffic

### v0.3.2 - November 1, 2024
Bugfixes, rework local and remote ends of switchport links

### v0.3.1 - October 29, 2024
Switch fixes: Hosts can link to switches, better broadcast handling, improved ARP + MAC tables

### v0.3.0 - October 26, 2024
User Achievements, ARP table, switch MAC address tables, user preferences

### v0.2.11 - October 25, 2024
Subnetting, DHCP + ping funcionality rework, generalize destination MAC finding

### v0.2.10 - September 30, 2024
Structs for messages and datagrams, sockets for internal application-layer communication

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

## License
This project is licensed under the terms of the GNU General Public License v3.0. See the [LICENSE](./LICENSE) file for details.