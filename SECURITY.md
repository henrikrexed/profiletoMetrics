# Security Policy

## Supported Versions

We release patches for security vulnerabilities in the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |
| < 0.1   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability, please report it to us privately:

1. **Email**: Send details to security@henrikrexed.com
2. **GitHub Security Advisory**: Use GitHub's private vulnerability reporting feature
3. **Include**:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

## Security Measures

This project implements several security measures:

- **Dependency Scanning**: Automated vulnerability scanning with Trivy and Gosec
- **Code Analysis**: Static analysis with CodeQL
- **OSS Scorecard**: Regular security assessments
- **Secure Dependencies**: Regular updates of dependencies
- **Container Security**: Multi-stage Docker builds with minimal attack surface

## Security Best Practices

When using this connector:

1. **Keep Dependencies Updated**: Regularly update OpenTelemetry Collector and dependencies
2. **Network Security**: Use TLS/SSL for OTLP endpoints in production
3. **Access Control**: Implement proper authentication and authorization
4. **Monitoring**: Monitor for unusual patterns in profiling data
5. **Configuration**: Review and validate all configuration files

## Security Considerations

- **Data Sensitivity**: Profiling data may contain sensitive information
- **Network Exposure**: Ensure OTLP endpoints are properly secured
- **Resource Limits**: Set appropriate resource limits for the collector
- **Logging**: Be cautious with debug logging in production environments

## Contact

For security-related questions or concerns, please contact:
- **Security Team**: security@henrikrexed.com
- **GitHub Issues**: Use the "security" label for public security discussions
