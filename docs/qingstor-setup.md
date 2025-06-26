# QingStor Configuration for Release Workflow

This document explains how to configure QingStor object storage for the automated release workflow.

## Overview

The release workflow automatically uploads release artifacts to QingStor object storage when a new tag is pushed. This requires proper configuration of access credentials.

## Required Secrets

To enable QingStor upload functionality, you need to add the following secrets to your GitHub repository:

### 1. KS_QSCTL_ACCESS_KEY_ID
- **Description**: QingStor access key ID
- **Value**: Your QingStor access key ID

### 2. KS_QSCTL_SECRET_ACCESS_KEY  
- **Description**: QingStor secret access key
- **Value**: Your QingStor secret access key

## How to Add Secrets

1. Go to your repository on GitHub
2. Navigate to **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Add each secret with the exact names listed above

### Direct Link
For this repository: https://github.com/kubesys/kubekey/settings/secrets/actions

## QingStor Account Setup

If you don't have QingStor credentials:

1. Sign up for a QingStor account at https://www.qingcloud.com/
2. Create an access key pair in the QingStor console
3. Note down the access key ID and secret access key
4. Ensure your account has permissions to write to the target bucket

## Bucket Configuration

The workflow uploads to the following QingStor locations:
- **Bucket**: `kubernetes`
- **Region**: `pek3b` 
- **Path Pattern**: `/kubekey/releases/download/{VERSION}/kubekey-{VERSION}-{OS}-{ARCH}.tar.gz`

Make sure your QingStor account has write permissions to this bucket.

## Workflow Behavior

### With Secrets Configured
- ✅ Artifacts are uploaded to both GitHub Releases and QingStor
- ✅ Full automation works as expected

### Without Secrets Configured  
- ✅ Artifacts are still uploaded to GitHub Releases
- ⚠️ QingStor upload is skipped with informational message
- ✅ Workflow continues and completes successfully

## Testing

To test the configuration:

1. Add the required secrets to your repository
2. Create and push a new tag: `git tag v0.1.2 && git push origin v0.1.2`
3. Check the workflow logs in the Actions tab
4. Verify artifacts appear in both GitHub Releases and QingStor

## Troubleshooting

### Common Issues

1. **"access key not provided"**
   - Check that secrets are named exactly: `KS_QSCTL_ACCESS_KEY_ID` and `KS_QSCTL_SECRET_ACCESS_KEY`
   - Verify secrets are added to the correct repository

2. **Permission denied errors**
   - Ensure your QingStor account has write permissions to the target bucket
   - Check that the access key is active and not expired

3. **Network connectivity issues**
   - QingStor endpoints may be blocked in some regions
   - Consider using a VPN or proxy if needed

### Debug Mode

To enable debug logging, add this to your workflow:

```yaml
- name: Debug QingStor Configuration
  run: |
    echo "Access Key ID length: ${#KS_QSCTL_ACCESS_KEY_ID}"
    echo "Secret Key length: ${#KS_QSCTL_SECRET_ACCESS_KEY}"
    qsctl ls qs://kubernetes/ -c qsctl-config.yaml
```

## Security Notes

- Never commit access keys to the repository
- Use GitHub Secrets to store sensitive credentials
- Regularly rotate your QingStor access keys
- Monitor access logs for unauthorized usage

## Support

For QingStor-specific issues:
- QingStor Documentation: https://docs.qingcloud.com/qingstor/
- QingStor Support: Contact QingCloud support

For workflow issues:
- Check GitHub Actions logs
- Review this documentation
- Open an issue in the repository
