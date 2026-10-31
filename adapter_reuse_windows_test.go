package tun

import (
	"fmt"
	"os"
	"testing"

	"github.com/sagernet/sing-tun/internal/wintun"
)

// This uses only the pre-existing public/native APIs, so it can also run
// unchanged against the unpatched parent revision.
func TestWintunDefaultReuse(t *testing.T) {
	if os.Getenv("SING_TUN_SYSTEM_TEST") != "1" {
		t.Skip("requires disposable administrator Windows runner")
	}
	name := fmt.Sprintf("tun-reuse-%d", os.Getpid())
	adapter, err := wintun.CreateAdapter(name, TunnelType, generateGUIDByDeviceName(name))
	if err != nil {
		t.Fatal(err)
	}
	defer adapter.Close()
	device, err := New(Options{Name: name, MTU: 1500, EXP_ExternalConfiguration: true})
	if err != nil {
		t.Fatalf("open previously created adapter: %T %#v (%v)", err, err, err)
	}
	defer device.Close()
	if actual, err := device.Name(); err != nil || actual != name {
		t.Fatalf("reused adapter: name=%q error=%v", actual, err)
	}
}
