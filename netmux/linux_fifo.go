// +build linux

package netmux

import "golang.org/x/sys/unix"

func mkfifo(path string, mode uint32) error {
	if err := unix.Mknod(path, unix.S_IFIFO|0666, 0); err != nil {
		return err
	}
	return nil
}
