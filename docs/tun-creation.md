# TUN creation regressions

These cases exercise the library without a proxy application, routes, upstream
servers, or a consumer-specific configuration. The original behavior is
reproducible at `040db71fcd33`.

| Case | Minimal setup | Failure before the fix | Expected behavior |
| --- | --- | --- | --- |
| PE section table | Allocate a small PE image in a Go byte slice; call Sections under race/checkptr | The giant array conversion claims memory beyond the allocation and terminates the process | Return exactly NumberOfSections entries within the image |
| Existing Wintun adapter | Create an adapter through the native API, then call New with the same name/GUID and default options | HRESULT 0x800700b7 is not recognized by errors.Is(err, os.ErrExist), so New never reaches OpenAdapter | Default reuse opens the existing adapter |
| Failed session creation | Inject StartSession and Close errors through the adapter boundary | Returning after StartSession fails leaves the acquired adapter open until a finalizer | Close once immediately and preserve both errors; retain ownership on success |

`EXP_Exclusive` is a separate opt-in behavior: callers that must own a newly
created device can reject existing Linux/Windows devices. Linux tests create a
persistent TUN without an active queue and check both exclusive rejection and
default attachment. Windows tests check exclusive rejection and default reuse.
This does not change the default attachment policy.

The existing `test.yml` platform jobs are retained. The normal Linux job adds the
exclusive-device case in an isolated namespace; one additional Windows Go 1.27.3
job runs PE/error/session/device regressions with race/checkptr. PRs to any base
branch can trigger these checks. Existing ping checks are unchanged.

Before-fix evidence is recorded in the [reproduction run](https://github.com/xjdrew/sing-tun/actions/runs/41204647630).
The exact PR base failed the PE test. With only the PE bounds fix applied so DLL
loading could proceed, the public existing-adapter case independently failed
with `0x800700b7`. Patched code passed both. The historical-base reproduction is
not repeated by every normal CI run.

Ordinary tests skip device creation. Privileged cases require
`SING_TUN_SYSTEM_TEST=1` on disposable administrator/root runners; Linux device
tests run in a fresh network namespace. Runtime race/checkptr stays enabled.

For a fixed revision, run on Windows:

```powershell
go test -race ./internal/wintun ./internal/wintun/memmod
$env:SING_TUN_SYSTEM_TEST='1'
go test -race -run '^Test(StartWintunSession|ExclusiveWintunCreation|WintunDefaultReuse)$' .
```

Root-package vet is checked. Full-module Windows vet also has pre-existing
unsafe-pointer diagnostics in native loaders and `%w` formatting diagnostics in
old winipcfg tests; these unrelated diagnostics are outside this patch.
