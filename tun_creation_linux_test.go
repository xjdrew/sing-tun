package tun

import (
	"errors"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/sagernet/netlink"
	"golang.org/x/sys/unix"
)

func TestExclusiveTunCreation(t *testing.T) {
	if os.Getenv("SING_TUN_SYSTEM_TEST") != "1" {
		t.Skip("requires disposable privileged network namespace")
	}
	name := fmt.Sprintf("tunfix%d", os.Getpid())
	fd, err := open(name, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := unix.IoctlSetInt(fd, unix.TUNSETPERSIST, 1); err != nil {
		unix.Close(fd)
		t.Fatal(err)
	}
	t.Cleanup(func() {
		link, err := netlink.LinkByName(name)
		if err == nil {
			if err := netlink.LinkDel(link); err != nil {
				t.Error(err)
			}
		}
	})
	if err := unix.Close(fd); err != nil {
		t.Fatal(err)
	}
	// A persistent device has no active queue. Default attachment is allowed,
	// so rejecting it specifically verifies IFF_TUN_EXCL rather than queue use.
	fd, err = open(name, false, true)
	if err == nil {
		unix.Close(fd)
	}
	if !errors.Is(err, unix.EBUSY) {
		t.Fatalf("exclusive creation must reject persistent interface: %v", err)
	}
	if _, err := net.InterfaceByName(name); err != nil {
		t.Fatalf("existing interface was lost: %v", err)
	}
	fd, err = open(name, false, false)
	if err != nil {
		t.Fatalf("default behavior must still attach persistent interface: %v", err)
	}
	if err := unix.Close(fd); err != nil {
		t.Fatal(err)
	}
}
