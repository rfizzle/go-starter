# Security Policy

## Supported versions

Go Starter is pre-1.0 and does not currently publish a long-term support
matrix. Security fixes are made on the default branch; older commits and
example deployments should be treated as unsupported. Consumers should track
the latest reviewed revision and maintain their own deployment-specific
hardening and dependency update process.

## Reporting a vulnerability

Do not disclose a suspected vulnerability, exploit, credential, or sensitive
log in a public issue, pull request, or discussion.

This repository does not currently publish a dedicated security email address
or other guaranteed private contact. First check the repository's GitHub
**Security** tab for a **Report a vulnerability** button and use it if
available. If GitHub private vulnerability reporting is unavailable, contact
the repository owner through their GitHub profile to request a private channel
without including vulnerability details. Share the details only after a
private channel has been confirmed.

Include, where safe:

- the affected revision or version;
- the component and prerequisites involved;
- reproducible steps or a minimal proof of concept;
- the impact and realistic attack scenario;
- any suggested mitigation;
- whether the issue is already public or actively exploited.

Maintainers should acknowledge the report through the private channel,
validate scope and severity, coordinate a fix and disclosure plan, and credit
the reporter if requested. No response or remediation SLA is promised at this
stage.

## Security expectations

The repository includes security controls and examples, but downstream
operators remain responsible for secrets, restrictive CORS settings, TLS
termination, database access, backups, monitoring, and network exposure. The
seeded example account and local-development credentials in the README must
never be used in production.

Run `make vuln` to scan Go dependencies and `make check` for the standard
validation gates. See the [deployment notes](README.md#deployment-notes) before
shipping a derived service.
