# GitHub Actions Security Workflow Troubleshooting

## 🚨 Common Issues and Solutions

### Issue: "Resource not accessible by integration"

**Error Message:**
```
Error: Resource not accessible by integration
```

**Cause:** Missing permissions for uploading security scan results to GitHub Security tab.

**Solution:** ✅ **FIXED** - Added required permissions to workflow:

```yaml
permissions:
  contents: read
  security-events: write
  deployments: write
```

### Issue: Trivy scan fails silently

**Symptoms:**
- Workflow passes but no SARIF file generated
- Security tab shows no results

**Solutions:**

1. **Check Trivy configuration:**
   ```yaml
   - name: Run Trivy vulnerability scanner
     uses: aquasecurity/trivy-action@master
     with:
       scan-type: 'fs'
       scan-ref: '.'
       format: 'sarif'
       output: 'trivy-results.sarif'
       severity: 'CRITICAL,HIGH,MEDIUM'
       config: '.github/security/trivy-config.yaml'  # ✅ Added
   ```

2. **Verify results file exists:**
   ```yaml
   - name: Check Trivy results
     run: |
       if [ -f "trivy-results.sarif" ]; then
         echo "Trivy scan completed successfully"
         ls -la trivy-results.sarif
       else
         echo "Trivy scan failed - no results file found"
         exit 1
       fi
   ```

### Issue: SARIF upload fails

**Error Message:**
```
Error: No SARIF files found
```

**Solutions:**

1. **Check file path:**
   ```yaml
   - name: Upload Trivy scan results
     uses: github/codeql-action/upload-sarif@v3
     if: always() && success()  # ✅ Added success check
     with:
       sarif_file: 'trivy-results.sarif'
       category: 'Trivy Security Scan'  # ✅ Added category
   ```

2. **Verify file generation:**
   ```bash
   # Check if file exists in workflow
   ls -la trivy-results.sarif
   
   # Check file content
   cat trivy-results.sarif | head -20
   ```

### Issue: Workflow permissions denied

**Error Message:**
```
Error: The workflow is not allowed to create check runs
```

**Solutions:**

1. **Repository Settings:**
   - Go to Settings → Actions → General
   - Ensure "Allow GitHub Actions to create and approve pull requests" is enabled
   - Check "Workflow permissions" section

2. **Organization Settings:**
   - If repository is in an organization, check organization policies
   - Ensure "Actions" are allowed for the repository

### Issue: Trivy database update fails

**Error Message:**
```
Error: failed to download vulnerability database
```

**Solutions:**

1. **Add retry logic:**
   ```yaml
   - name: Run Trivy vulnerability scanner
     uses: aquasecurity/trivy-action@master
     with:
       scan-type: 'fs'
       scan-ref: '.'
       format: 'sarif'
       output: 'trivy-results.sarif'
       severity: 'CRITICAL,HIGH,MEDIUM'
       config: '.github/security/trivy-config.yaml'
       retry: 3  # ✅ Add retry attempts
   ```

2. **Use cached database:**
   ```yaml
   - name: Run Trivy vulnerability scanner
     uses: aquasecurity/trivy-action@master
     with:
       scan-type: 'fs'
       scan-ref: '.'
       format: 'sarif'
       output: 'trivy-results.sarif'
       severity: 'CRITICAL,HIGH,MEDIUM'
       config: '.github/security/trivy-config.yaml'
       skip-db-update: true  # ✅ Skip DB update if recent
   ```

## 🔧 Debugging Steps

### 1. Enable Debug Logging

Add this to your workflow:

```yaml
env:
  ACTIONS_STEP_DEBUG: true
  TRIVY_DEBUG: true
```

### 2. Check Workflow Logs

```bash
# View workflow run
gh run view <run-id>

# Download logs
gh run download <run-id>

# View specific job logs
gh run view <run-id> --log
```

### 3. Test Locally

```bash
# Install Trivy
brew install trivy  # macOS
# or
sudo apt-get install trivy  # Ubuntu

# Run scan locally
trivy fs --format sarif --output trivy-results.sarif .

# Verify SARIF file
cat trivy-results.sarif | jq .
```

## 📊 Expected Results

### Successful Workflow

1. **Trivy Scan**: ✅ Completes without errors
2. **Results File**: ✅ `trivy-results.sarif` generated
3. **Upload**: ✅ Results appear in GitHub Security tab
4. **Security Alerts**: ✅ Vulnerabilities shown as Code scanning alerts

### Security Tab Location

- **Repository**: Security → Code scanning alerts
- **Organization**: Security → Code scanning → [Repository Name]
- **Enterprise**: Security → Code scanning → [Organization] → [Repository]

## 🚀 Performance Optimization

### 1. Cache Trivy Database

```yaml
- name: Cache Trivy DB
  uses: actions/cache@v3
  with:
    path: ~/.cache/trivy
    key: ${{ runner.os }}-trivy-${{ hashFiles('go.mod', 'go.sum') }}
    restore-keys: |
      ${{ runner.os }}-trivy-
```

### 2. Parallel Security Scans

```yaml
jobs:
  security-fs:
    name: Filesystem Security Scan
    runs-on: ubuntu-latest
    # ... filesystem scan

  security-deps:
    name: Dependencies Security Scan
    runs-on: ubuntu-latest
    # ... dependency scan
```

### 3. Conditional Scanning

```yaml
- name: Run Trivy vulnerability scanner
  uses: aquasecurity/trivy-action@master
  if: github.event_name == 'push' || github.event_name == 'pull_request'
  with:
    scan-type: 'fs'
    scan-ref: '.'
    format: 'sarif'
    output: 'trivy-results.sarif'
```

## 📞 Getting Help

### GitHub Support

- **Documentation**: [GitHub Actions](https://docs.github.com/en/actions)
- **Community**: [GitHub Community](https://github.community/)
- **Support**: [GitHub Support](https://support.github.com/)

### Trivy Support

- **Documentation**: [Trivy Docs](https://aquasecurity.github.io/trivy/)
- **Issues**: [GitHub Issues](https://github.com/aquasecurity/trivy/issues)
- **Discussions**: [GitHub Discussions](https://github.com/aquasecurity/trivy/discussions)

---

**Last Updated**: January 2025  
**Workflow Version**: 2.0  
**Status**: ✅ Security scanning fully configured
