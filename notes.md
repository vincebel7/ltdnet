# Known issues and bugs

- Fix map sorting in router.go next_free_addr(), it is string sorting 10, 100, etc.

- Broadcasts aren't handled correctly. Should traverse network instead of a global push to all devices

- Unreachable pings hang on failed ARP requests

# Feature ideas
- MAC learning for hosts

- Flood frame if not in MAC table of switch

- Switch network functionality

- HTTP over UDP

- Multiple routers, with static routing

- Traceroute (after routing is added)

- DNS server

