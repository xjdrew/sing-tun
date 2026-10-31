package tun

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/sagernet/sing-tun/internal/wintun"
)

type failingSessionAdapter struct {
	startErr, closeErr error
	closed             int
}

func (a *failingSessionAdapter) StartSession(_ uint32) (wintun.Session, error) {
	return wintun.Session{}, a.startErr
}

func (a *failingSessionAdapter) Close() error {
	a.closed++
	return a.closeErr
}

func TestStartWintunSession(t *testing.T) {
	startErr := errors.New("injected session failure")
	closeErr := errors.New("injected adapter close failure")
	for _, failure := range []bool{false, true} {
		t.Run(fmt.Sprint(failure), func(t *testing.T) {
			a := &failingSessionAdapter{closeErr: closeErr}
			if failure {
				a.startErr = startErr
			}
			_, err := startWintunSession(a)
			if failure {
				if !errors.Is(err, startErr) || !errors.Is(err, closeErr) || a.closed != 1 {
					t.Fatalf("failed session must close adapter and retain errors: closed=%d, error=%v", a.closed, err)
				}
			} else if err != nil || a.closed != 0 {
				t.Fatalf("successful session must retain adapter: closed=%d, error=%v", a.closed, err)
			}
		})
	}
}

func TestExclusiveWintunCreation(t *testing.T) {
	if os.Getenv("SING_TUN_SYSTEM_TEST") != "1" {
		t.Skip("requires disposable administrator Windows runner")
	}
	options := Options{Name: fmt.Sprintf("tun-ci-%d", os.Getpid()), MTU: 1500, EXP_ExternalConfiguration: true, EXP_Exclusive: true}
	first, err := New(options)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := New(options)
	if second != nil {
		second.Close()
	}
	if !errors.Is(err, os.ErrExist) {
		t.Fatalf("exclusive creation must reject existing adapter: %T %#v (%v)", err, err, err)
	}
	if name, err := first.Name(); err != nil || name != options.Name {
		t.Fatalf("original adapter changed: %q %v", name, err)
	}
}
