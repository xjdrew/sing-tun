package wintun

import (
	"errors"
	"os"
	"syscall"
	"testing"
)

func TestAdapterErrorNormalizesWin32HRESULT(t *testing.T) {
	for _, tc := range []struct {
		code syscall.Errno
		want error
	}{
		{183, os.ErrExist},
		{0x800700b7, os.ErrExist},
		{0x80070005, os.ErrPermission},
		{0x80070002, os.ErrNotExist},
		{0x80004005, syscall.Errno(0x80004005)},
	} {
		if err := adapterError(tc.code); !errors.Is(err, tc.want) {
			t.Errorf("error %#x became %v, want %v", tc.code, err, tc.want)
		}
	}
}
