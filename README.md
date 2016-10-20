# netmux
Take a TCP or UDP connection and broadcast it to other TCP/UDP listeners, Unix named pipes, and Unix domain sockets.

If a named pipe in the config does not exist, `netmux` will try to create it,
 but it won't overwrite existing named pipes.

## Usage
    go install github.com/graysonchao/netmux
    netmux --config <config.json>

An example configuration is provided. It includes one of each type of output:

* "fifo" is a named pipe at `/usr/local/var/netmux.1`
* "udp0" sends output over UDP to 127.0.0.1:8889.
* "tcp0" sends output over TCP to 127.0.0.1:8890
* "my_unix" sends output to a Unix domain socket at `/usr/local/var/netmux.sock`.

If you try to use the example config, you should create the Unix domain socket first:

    nc -U -l /usr/local/var/netmux.sock

## Notes
`netmux` uses `unix.mkfifo` on FreeBSD and Darwin and `unix.mknod` on Linux.

The Unix domain socket type used is SOCK_STREAM.
