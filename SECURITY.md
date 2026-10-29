# Security Policy

## Supported Versions

We release patches for security vulnerabilities. Which versions are eligible for receiving such patches depends on the CVSS v3.0 Rating:

| Version | Supported          |
| ------- | ------------------ |
| 1.1.x   | :white_check_mark: |
| 1.0.x   | :x:                |

Build with **Go 1.26.10+** or **1.27.3+**. Go 1.25 and older are end of life and do not receive the October 2026 security releases. Those fixes are in the toolchain (`crypto/x509`, `net/http`, `html/template`), not in this module. The `go` line asks the `go` command for at least Go 1.26.10 when `GOTOOLCHAIN=auto`, which is the default. `GOTOOLCHAIN=local` fails the build until that toolchain is installed.

- [CVE-2026-111943](https://pkg.go.dev/vuln/GO-2026-6809) (`crypto/x509`): untrustworthy eMudhra intermediates are rejected. Fixed in Go 1.26.10 and 1.27.3 ([announcement](https://groups.google.com/g/golang-announce/c/66pojueaQek/m/tmWX2Pq9AgAJ)).
- Go 1.26.9 and 1.27.2 also fixed HTTP/2 server crashes, trailer memory use, and `html/template` context tracking ([announcement](https://groups.google.com/g/golang-announce/c/U2fTuyDJznI)).
- This server does not import `github.com/andybalholm/brotli`. Versions before v1.2.7 panic on crafted Brotli input ([CVE-2026-111497](https://www.wiz.io/blog/cve-2026-111497-brotli-rce-node-js)). Any future import must be v1.2.7 or newer. Gin is required at v1.12.0.
- `golang.org/x/net` v0.60.0 and `github.com/quic-go/quic-go` v0.59.1 close HTTP/2 and HTTP/3 issues that Gin pulls into HTTP handlers.

## Reporting a Vulnerability

The go-chess team and community take security bugs seriously. We appreciate your efforts to responsibly disclose your findings, and will make every effort to acknowledge your contributions.

To report a security issue, please use the GitHub Security Advisory ["Report a Vulnerability"](https://github.com/rumendamyanov/go-chess/security/advisories/new) tab.

The go-chess team will send a response indicating the next steps in handling your report. After the initial reply to your report, the security team will keep you informed of the progress towards a fix and full announcement, and may ask for additional information or guidance.

### When Should I Report a Vulnerability?

- You think you discovered a potential security vulnerability in go-chess
- You are unsure how a vulnerability affects go-chess
- You think you discovered a vulnerability in another project that go-chess depends on

### When Should I NOT Report a Vulnerability?

- You need help tuning go-chess components for security
- You need help applying security related updates
- Your issue is not security related

## Security Update Process

If a security vulnerability is reported, the following process will be followed:

1. **Confirmation**: We'll work to confirm the vulnerability
2. **Assessment**: We'll assess the impact and assign a severity level
3. **Fix Development**: We'll develop and test a fix
4. **Disclosure**: We'll coordinate disclosure with the reporter
5. **Release**: We'll release a patched version
6. **Announcement**: We'll announce the vulnerability and fix

## Preferred Languages

We prefer all communications to be in English.

## Comments on This Policy

If you have suggestions on how this process could be improved please submit a pull request.
