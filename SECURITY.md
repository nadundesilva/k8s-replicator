# Security Policy 🔒

Security is our top priority! 🛡️ This document outlines our security practices, vulnerability reporting process, and security features for K8s Replicator.

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.6.x   | :white_check_mark: |
| 0.5.x   | :x:                |
| < 0.5   | :x:                |

## Reporting a Vulnerability 🚨

**DO NOT** create public GitHub issues for security vulnerabilities.

**Include:**

- Description of the vulnerability
- Steps to reproduce
- Potential impact assessment
- Suggested fix (if any)
- Your contact information

**Response Timeline:**

- Initial Response: Within 48 hours
- Status Update: Within 7 days
- Resolution: Within 30 days

## Security Best Practices 🛡️

**For Users:**

- Use latest stable version
- Update dependencies regularly
- Review RBAC permissions
- Monitor logs for suspicious activity
- Use network policies to restrict access

**For Contributors:**

- Follow secure coding practices
- Run security scans
- Keep dependencies updated
- Review security implications of changes

## Security Features 🔐

**Built-in Security:**

- RBAC Integration
- Namespace Isolation
- Label-based Filtering
- Resource Validation

**Security Considerations:**

- Secret Management with encryption
- Network Policy security boundaries
- Audit Logging
- Resource Quota compliance

## Security Advisories 📢

Published in:

- [GitHub Security Advisories](https://github.com/nadundesilva/k8s-replicator/security/advisories)
- [CHANGELOG.md](CHANGELOG.md)
- [Releases](https://github.com/nadundesilva/k8s-replicator/releases)

## Security Contacts 📞

- **Maintainer**: [@nadundesilva](https://github.com/nadundesilva)
- **GitHub Security**: Use GitHub's private vulnerability reporting

---

**Note**: This security policy is subject to change. Please check back regularly for updates.

Thank you for helping keep K8s Replicator secure! 🛡️💙
