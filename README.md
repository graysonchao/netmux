# udpmux
Take a UDP connection and broadcast it to other UDP listeners, Unix named pipes, and Unix domain sockets.

If a named pipe in the config does not exist, `udpmux` will try to create it,
 but it won't overwrite existing named pipes.

## Usage
    go install github.com/graysonchao/udpmux
    udpmux --config <config.json>

An example configuration is provided. It includes one of each type of output:

* "fifo" is a named pipe at `/usr/local/var/udpmux.1`
* "udp0" sends output over UDP to 127.0.0.1:8889.
* "my_unix" sends output to a Unix domain socket at `/usr/local/var/udpmux.sock`.

If you try to use the example config, you should create the Unix domain socket first:

    nc -U -l /usr/local/var/udpmux.sock

## Notes
`udpmux` uses `unix.mkfifo` on FreeBSD and Darwin and `unix.mknod` on Linux.

The Unix domain socket type used is SOCK_STREAM.
