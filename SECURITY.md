# Security Policy

## Supported Versions
The project is under active development. All commits to `main` and `dev` branches undergo:
- Static analysis (golangci-lint, staticcheck)
- Dependency vulnerability scanning (govulncheck)
- CodeQL security scanning

## Reporting a Vulnerability

If you discover a security issue, please report it responsibly using this repository’s official GitHub security advisory workflow:

Navigate to: **Security → Report a vulnerability**

GitHub will create a private security report that is visible only to the project maintainers.  
We will review, triage, and respond as quickly as possible.


## Security Principles
pvecli follows strict security practices:
- No plaintext passwords
- Token-based authentication only
- All traffic uses TLS (API communication)
- SSH host key verification via known_hosts
- No remote code execution inside VMs or containers
- Minimal, vetted dependencies
