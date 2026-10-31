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

The existing `test.yml` workflow runs normal platform tests and the new race
regressions on supported Go versions. A Windows PR job also copies the first two
tests to the exact PR base: the PE case must either pass or fail with the specific
checkptr diagnostic. It then applies only the PE fix so native DLL loading can
proceed and checks the existing-adapter case independently. Unexpected build,
driver or environment errors fail the job. Results are retained in job logs.

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
