# Security Policy

## Supported Versions

| Version | Supported |
|---------|------------|
| Current main branch | ✅ Yes |
| Previous releases | ❌ No |

## Reporting a Vulnerability

If you discover a security vulnerability in OpenSustain, please report it responsibly.

### How to Report

**Do NOT** open a public issue for security vulnerabilities.

Instead, please send an email to: security@opensustain.org

Include the following information in your report:
- Description of the vulnerability
- Steps to reproduce the issue
- Potential impact of the vulnerability
- Any suggested mitigation or fix (if known)

### What to Expect

- We will acknowledge receipt of your report within 48 hours
- We will provide a detailed response within 7 days
- We will work with you to understand and validate the issue
- We will coordinate a fix and release schedule
- We will credit you in the release notes (unless you prefer anonymity)

### Security Best Practices for Users

When using OpenSustain:

1. **GitHub Tokens**: 
   - Never hardcode GitHub tokens in your code or configuration files
   - Use GitHub Secrets or environment variables
   - Use the minimum required token scopes (`read:org`, `repo` for org scans)
   - Rotate tokens regularly

2. **Docker Images**:
   - Use official Docker images from trusted sources
   - Keep Docker images updated
   - Scan images for vulnerabilities before deployment

3. **Local Scans**:
   - Be cautious when scanning repositories with sensitive data
   - Review generated reports before sharing them
   - Ensure report files have appropriate permissions

4. **CI/CD Integration**:
   - Use GitHub's built-in `GITHUB_TOKEN` when possible
   - Limit token scopes to minimum required permissions
   - Review workflow files for security best practices

## Security Features

OpenSustain includes the following security considerations:

- No hardcoded credentials or API keys
- Graceful handling of missing authentication
- Token validation before API calls
- Minimal required permissions for GitHub API access
- Static binary compilation for reproducible builds

## Dependency Management

We regularly update dependencies to address security vulnerabilities. Dependabot is enabled to automatically monitor and alert on dependency updates.

## License

This security policy is part of the OpenSustain project and is licensed under the MIT License.
