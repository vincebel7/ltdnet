# Known issues and bugs

- DHCP table/DHCP index a bit redundant/wasteful. Restructure, do this before MAC learning for hosts (same code)

- Broadcasts aren't handled correctly. Should traverse network instead of a global push to all devices

- With MAC address rework, unreachable pings hang on failed ARP requests

# Feature ideas
- MAC learning for hosts

- Flood frame if not in MAC table of switch

- Switch network functionality

- HTTP over UDP

- Communicate with running machine somehow (WAN)

- Multiple routers, with static routing

- Traceroute (after routing is added)
