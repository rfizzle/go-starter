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

## Triaging dependency vulnerabilities

CI runs the pinned `govulncheck` version for every proposed change, and the
scheduled Security workflow repeats the scan weekly. When a scan or Dependabot
alert reports a vulnerability:

1. Reproduce it from the affected revision with `make vuln`.
2. Record the vulnerability ID, affected module and version, fixed version,
   reported symbol, and whether `govulncheck` found a reachable call path.
3. Confirm the advisory and fix in the Go vulnerability database or the
   upstream project's security notice. Assess realistic prerequisites, exposed
   boundaries, and impact; do not dismiss an unreachable finding without a
   recorded rationale.
4. Prefer the smallest compatible dependency update that fixes the issue.
   Treat major-version updates as separate reviewed changes rather than
   combining them with routine minor and patch updates.
5. Run `make vuln`, `make check`, and any focused or race tests for affected
   code before merging. Keep reachable vulnerabilities out of releases until
   they are fixed or a time-bounded mitigation and follow-up are documented.

If no fix is available, document the affected deployments, compensating
controls, owner, review date, and the condition that will remove the
exception. Do not suppress a scanner result globally to handle one finding.

## Security expectations

The repository includes security controls and examples, but downstream
operators remain responsible for secrets, restrictive CORS settings, TLS
termination, database access, backups, monitoring, and network exposure. The
seeded example account and local-development credentials in the README must
never be used in production.

Run `make vuln` to scan Go dependencies and `make check` for the standard
validation gates. See the [deployment notes](README.md#deployment-notes) before
shipping a derived service.
