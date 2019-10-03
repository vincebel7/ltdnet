#Known issues and bugs
-Pings don't time out, leaving program hanging when pinging unknown address
-Program requires restart between creating host and being able to DHCP. Saving/reloading doesn't fix this, not sure why
	-Does this have to do with Listener not updating?
-Concurrent read/write error upon loading file
