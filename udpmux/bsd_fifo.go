// +build darwin freebsd

package udpmux

import "golang.org/x/sys/unix"

func mkfifo(path string, mode uint32) error {
	if err := unix.Mkfifo(path, 0666); err != nil {
		return err
	}
	return nil
}
