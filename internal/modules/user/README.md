# User module

This package is the starter's example business module and PostgreSQL repository
adapter. Shared user data, inputs, errors, and the repository contract live in
`internal/domain`; pgx and sqlc details stay in this package.

The package owns user persistence and transaction scoping. It does not own
authentication, password hashing, HTTP mapping, pool lifecycle, or migrations.
The application and seed command construct `Repository`; other consumers depend
on `domain.UserRepository` or a narrower consumer-side interface.

See [user.spec.md](user.spec.md) for the behavioral contract and
[the module/platform decision](../../../docs/architecture/module-platform-boundaries.md)
for dependency rules.
