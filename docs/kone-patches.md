# KONE compatibility fixes

This maintenance branch is based on upstream commit `040db71fcd33`
(`v0.9.7-0.20261023115432-040db71fcd33`), used by sing-box v1.14.4.
It deliberately excludes unrelated changes on the upstream development branch.
The module path stays `github.com/sagernet/sing-tun`; consumers can replace it
with a pinned commit from `github.com/xjdrew/sing-tun`.

- Bound the PE section-table slice to its actual count. The previous giant
  array cast fails Go race/checkptr when the PE image is backed by a Go allocation.
- Close the acquired Wintun adapter immediately when session creation fails;
  preserve both session and cleanup errors, without relying on a finalizer.
- Add opt-in `Options.EXP_Exclusive`: Linux uses `IFF_TUN_EXCL`, and Windows
  does not reopen an existing adapter after `CreateAdapter` reports it exists.
  The default attachment behavior remains available to existing library users.

PR CI runs Go 1.25 and 1.27.3 race regressions on Linux and Windows. It tests
heap-backed PE section tables, injected session/cleanup failures, real Wintun
duplicate creation, and persistent Linux TUN attachment in an isolated network
namespace. Privileged tests require `SING_TUN_SYSTEM_TEST=1` on disposable runners.
KONE additionally verifies real DNS/TCP/UDP traffic, gateway routing and rollback
with system/mixed stacks against the pinned fork commit.

CI vets the root package. Full-module Windows `go vet ./...` also reports existing
unsafe-pointer conversions in the native loaders and `%w` formatting in old
winipcfg tests; these unrelated baseline diagnostics are outside this patch.
The PE loader regression keeps runtime race/checkptr enabled.
