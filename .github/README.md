# GitHub Workflows

This directory contains GitHub Actions workflows for the ProfileToMetrics project.

## Workflows Overview

### 1. CI/CD Pipeline (`.github/workflows/ci.yml`)
**Main workflow that runs on every push and pull request**

- **Build and Test**: Multi-platform testing with Go 1.23 and 1.24
- **Security Scan**: Gosec and Trivy vulnerability scanning
- **OSS Scorecard**: Open Source Security Scorecard analysis
- **Docker Build**: Multi-platform Docker image building
- **Integration Tests**: End-to-end testing with real data
- **Code Quality**: Linting, formatting, and coverage checks
- **Documentation**: API documentation generation
- **Release**: Automated release creation

### 2. Security Scan (`.github/workflows/security.yml`)
**Comprehensive security scanning**

- **Dependency Scan**: Go vulnerability checking with `govulncheck`
- **CodeQL Analysis**: Static code analysis for security vulnerabilities
- **OSS Scorecard**: Weekly security scorecard analysis
- **Security Headers**: Check for security-related files and patterns
- **License Check**: License compliance verification

### 3. Docker Build (`.github/workflows/docker.yml`)
**Docker image building and testing**

- **Multi-platform Build**: Linux AMD64 and ARM64 support
- **Multi-version Testing**: Go 1.23 and 1.24 support
- **Security Scan**: Trivy vulnerability scanning of Docker images
- **Performance Test**: Image size and startup time testing
- **Functionality Test**: OTLP endpoint testing

### 4. Code Quality (`.github/workflows/code-quality.yml`)
**Code quality and formatting checks**

- **Format Check**: Go formatting and vetting
- **Linting**: golangci-lint, staticcheck, ineffassign, misspell, gocyclo
- **Documentation**: API documentation checks
- **Test Coverage**: Coverage reporting with 80% threshold
- **Performance**: Benchmark testing

### 5. Connector Tests (`.github/workflows/connector-test.yml`)
**Specific testing for the ProfileToMetrics connector**

- **Configuration Tests**: Test different configuration files
- **Data Tests**: Test with real trace, log, and metrics data
- **Performance Tests**: Benchmark and memory profiling
- **Docker Tests**: Test connector in Docker environment

### 6. OSS Scorecard (`.github/workflows/oss-scorecard.yml`)
**Open Source Security Scorecard analysis**

- **Scorecard Analysis**: Weekly security scorecard
- **Security Policy**: Check for security documentation
- **License Compliance**: License file and compatibility checks
- **Documentation Quality**: README, CONTRIBUTING, CHANGELOG checks

### 7. Release (`.github/workflows/release.yml`)
**Automated release management**

- **Release Creation**: Automated release from tags
- **Docker Publishing**: Multi-platform Docker image publishing
- **Asset Upload**: Binary releases for multiple platforms
- **Changelog Generation**: Automatic changelog from git commits

### 8. Dependabot (`.github/workflows/dependabot.yml`)
**Automated dependency updates**

- **Auto-merge**: Automatically merge minor and patch updates
- **Testing**: Run tests before auto-merging
- **Target**: Minor version updates only

## Configuration Files

### Dependabot Configuration (`.github/dependabot.yml`)
- **Go Modules**: Weekly updates for Go dependencies
- **GitHub Actions**: Weekly updates for GitHub Actions
- **Docker**: Weekly updates for Docker base images
- **Auto-assignment**: Automatic assignment to maintainers

## Security Features

### Vulnerability Scanning
- **Gosec**: Go security scanner
- **Trivy**: Container and filesystem vulnerability scanner
- **CodeQL**: Static analysis for security vulnerabilities
- **OSS Scorecard**: Open source security scorecard

### Security Policies
- **SECURITY.md**: Security vulnerability reporting
- **LICENSE**: MIT License compliance
- **CONTRIBUTING.md**: Contribution guidelines
- **CHANGELOG.md**: Change tracking

## Quality Gates

### Code Quality
- **Formatting**: Go fmt compliance
- **Linting**: Multiple linters (golangci-lint, staticcheck, etc.)
- **Coverage**: 80% test coverage threshold
- **Documentation**: API documentation generation

### Testing
- **Unit Tests**: Comprehensive unit test coverage
- **Integration Tests**: End-to-end testing
- **Performance Tests**: Benchmark and profiling
- **Docker Tests**: Container functionality testing

### Security
- **Dependency Scanning**: Regular vulnerability checks
- **Code Analysis**: Static security analysis
- **OSS Scorecard**: Weekly security scoring
- **License Compliance**: License verification

## Usage

### Running Workflows Locally
```bash
# Run tests
make test

# Run linting
make lint

# Run security scan
make security-scan

# Build Docker image
make docker-build
```

### Manual Workflow Triggers
- **workflow_dispatch**: Manual trigger for all workflows
- **Release**: Tag-based release workflow
- **Schedule**: Weekly security and scorecard checks

### Workflow Dependencies
```
CI/CD Pipeline
├── Build and Test
├── Security Scan
├── OSS Scorecard
├── Docker Build
├── Integration Tests
├── Code Quality
├── Documentation
└── Release (on tags)
```

## Monitoring

### Status Badges
Add these badges to your README:

```markdown
![CI/CD](https://github.com/henrikrexed/profiletoMetrics/workflows/CI/CD%20Pipeline/badge.svg)
![Security](https://github.com/henrikrexed/profiletoMetrics/workflows/Security%20Scan/badge.svg)
![Docker](https://github.com/henrikrexed/profiletoMetrics/workflows/Docker%20Build%20and%20Push/badge.svg)
![OSS Scorecard](https://github.com/henrikrexed/profiletoMetrics/workflows/OSS%20Scorecard/badge.svg)
```

### Security Tab
- **Code Scanning**: CodeQL and security scan results
- **Dependabot**: Dependency vulnerability alerts
- **Secret Scanning**: Secret detection and alerts
- **OSS Scorecard**: Security scorecard results

## Troubleshooting

### Common Issues
1. **Build Failures**: Check Go version compatibility
2. **Test Failures**: Verify test data and configuration
3. **Security Alerts**: Review and fix vulnerability reports
4. **Docker Issues**: Check platform compatibility

### Debug Information
- **Logs**: Available in GitHub Actions logs
- **Artifacts**: Coverage reports and documentation
- **Security**: Security scan results in Security tab
- **Scorecard**: OSS Scorecard results in Security tab

## Contributing

### Adding New Workflows
1. Create workflow file in `.github/workflows/`
2. Follow existing patterns and naming conventions
3. Add appropriate permissions and triggers
4. Test locally before committing

### Modifying Existing Workflows
1. Test changes in a fork or branch
2. Verify all dependent workflows still work
3. Update documentation if needed
4. Submit pull request with clear description

## Support

For issues with workflows:
1. Check GitHub Actions logs
2. Review workflow configuration
3. Test locally with `act` or similar tools
4. Create issue with detailed information
