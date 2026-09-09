# Local Symphony license library patch

Source: github.com/symphonylicensemanager/go v0.1.0 (module, license and machineid packages).

Windows HypervisorPresent also describes physical hosts with an active hypervisor.
The Windows VM detector now uses guest model markers instead, allowing physical
HP Victus hosts while retaining guest VM detection. The machine ID hashing,
embedded public key, signature verification, expiry and machine matching are unchanged.

go.mod uses a local replace so both machineid.ID and license.Verify share this fix.
Run `go test ./...` from this directory to test the patched dependency.