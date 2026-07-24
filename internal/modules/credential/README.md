# Credential module

This package verifies a plaintext password against a user's stored bcrypt
hash. It is the starter's example authentication business capability,
independent of HTTP, generated transport, and any concrete storage.

The package owns password verification only. It does not hash passwords for
storage (`internal/cli/seed` and any future signup flow own that), issue
tokens or cookies, or read/write any user field beyond the
authentication-only projection it consumes.

`Service` depends on a package-private `credentialReader` port, satisfied
structurally by `internal/modules/user.Repository`'s `CredentialByEmail`
method. A future composition root constructs the concrete repository and
injects it; this package never imports `internal/modules/user`.

See [credential.spec.md](credential.spec.md) for the behavioral contract and
[the module/platform decision](../../../docs/architecture/module-platform-boundaries.md)
for dependency rules.
