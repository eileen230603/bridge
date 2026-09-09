# Local Symphony license library patch

Source: github.com/symphonylicensemanager/go v0.1.0 (module, license and machineid packages).

Windows HypervisorPresent also describes physical hosts with an active hypervisor.
The Windows VM detector now uses guest model markers instead, allowing physical
HP Victus hosts while retaining guest VM detection. The machine ID hashing,
signature verification, expiry and machine matching are unchanged. The public key was replaced for the project-owned issuer.

go.mod uses a local replace so both machineid.ID and license.Verify share this fix.
Run `go test ./...` from this directory to test the patched dependency.
The issuer public key now belongs to this project; upstream license tokens are no longer accepted. See docs/license-admin.md in the repository root.
