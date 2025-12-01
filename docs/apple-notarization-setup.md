# Apple Notarization Setup Guide

This guide walks through setting up code signing and notarization for macOS binaries.

## Prerequisites

- Apple Developer Account ($99/year) - https://developer.apple.com/programs/
- Access to a Mac with Xcode Command Line Tools installed
- Admin access to GitHub repository settings

## Step 1: Create Developer ID Application Certificate

1. Go to https://developer.apple.com/account/resources/certificates
2. Click the **+** button to create a new certificate
3. Select **Developer ID Application**
4. Follow the instructions to create a Certificate Signing Request (CSR):
   ```bash
   # Open Keychain Access
   # Keychain Access > Certificate Assistant > Request a Certificate from a Certificate Authority
   # Fill in your email, name, and select "Saved to disk"
   ```
5. Upload the CSR file
6. Download the certificate (*.cer file)
7. Double-click to import into Keychain Access

## Step 2: Export Certificate as .p12

1. Open **Keychain Access**
2. Find your "Developer ID Application" certificate under "My Certificates"
3. Expand it to see both the certificate and private key
4. Right-click on the certificate and select **Export**
5. Choose **.p12** format
6. **Set a strong password** (you'll need this later)
7. Save as `DeveloperID_Application.p12`

## Step 3: Create App Store Connect API Key

1. Go to https://appstoreconnect.apple.com/access/api
2. Click the **+** button under "Keys"
3. Enter a name (e.g., "Clem GoReleaser")
4. Select **Developer** role
5. Click **Generate**
6. **Download the .p8 file immediately** (only available once!)
7. Note the **Key ID** (looks like: `ABC123DEFG`)
8. Note the **Issuer ID** (looks like: `12345678-1234-1234-1234-123456789abc`)

## Step 4: Base64 Encode Files

On your Mac, encode both files:

```bash
# Encode the .p12 certificate
base64 -i DeveloperID_Application.p12 | pbcopy
# Paste this into a text file as MACOS_SIGN_P12.txt

# Encode the .p8 API key
base64 -i AuthKey_ABC123DEFG.p8 | pbcopy
# Paste this into a text file as MACOS_NOTARY_KEY.txt
```

## Step 5: Add GitHub Secrets

Go to your repository settings:
https://github.com/2389-research/clem/settings/secrets/actions

Add these 5 secrets:

### 1. MACOS_SIGN_P12
- **Value**: The base64 string from `MACOS_SIGN_P12.txt`
- **Description**: Base64-encoded Developer ID Application certificate

### 2. MACOS_SIGN_PASSWORD
- **Value**: The password you set when exporting the .p12 file
- **Description**: Password for the .p12 certificate

### 3. MACOS_NOTARY_KEY
- **Value**: The base64 string from `MACOS_NOTARY_KEY.txt`
- **Description**: Base64-encoded App Store Connect API key

### 4. MACOS_NOTARY_KEY_ID
- **Value**: The Key ID from App Store Connect (e.g., `ABC123DEFG`)
- **Description**: App Store Connect API Key ID

### 5. MACOS_NOTARY_ISSUER_ID
- **Value**: The Issuer ID from App Store Connect (e.g., `12345678-1234-1234-1234-123456789abc`)
- **Description**: App Store Connect Issuer ID

## Step 6: Update GitHub Actions Workflow

The workflow needs to pass these secrets to GoReleaser:

```yaml
- name: Run GoReleaser
  uses: goreleaser/goreleaser-action@v6
  with:
    distribution: goreleaser
    version: latest
    args: release --clean
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    MACOS_SIGN_P12: ${{ secrets.MACOS_SIGN_P12 }}
    MACOS_SIGN_PASSWORD: ${{ secrets.MACOS_SIGN_PASSWORD }}
    MACOS_NOTARY_KEY: ${{ secrets.MACOS_NOTARY_KEY }}
    MACOS_NOTARY_KEY_ID: ${{ secrets.MACOS_NOTARY_KEY_ID }}
    MACOS_NOTARY_ISSUER_ID: ${{ secrets.MACOS_NOTARY_ISSUER_ID }}
```

## Step 7: Test the Configuration

Once secrets are configured:

1. Create a test tag: `git tag v1.0.1-test && git push origin v1.0.1-test`
2. Watch the GitHub Actions workflow
3. If successful, the binaries will be:
   - Signed with your Developer ID
   - Notarized by Apple
   - Ready to run without Gatekeeper warnings

## Verification

Download a binary from the release and verify:

```bash
# Check code signature
codesign -dv --verbose=4 clem

# Should show:
# Authority=Developer ID Application: Your Name (TEAM_ID)
# TeamIdentifier=YOUR_TEAM_ID

# Check notarization
spctl -a -vv -t install clem

# Should show:
# clem: accepted
# source=Notarized Developer ID
```

## Troubleshooting

### "Certificate not found" error
- Verify the .p12 file was base64 encoded correctly
- Make sure you exported the certificate WITH the private key

### "Invalid API key" error
- Double-check the Key ID and Issuer ID match exactly
- Ensure the .p8 file was base64 encoded correctly
- Verify the API key has "Developer" role

### "Password incorrect" error
- Confirm the password matches what you set during .p12 export
- Try exporting a new .p12 with a different password

### Notarization takes a long time
- Notarization can take 5-15 minutes for Apple's servers
- GoReleaser will wait and poll for completion
- Check logs for status updates

## Security Best Practices

1. **Never commit** the .p12, .p8, or base64 files to git
2. **Store backups** of your certificates in a secure location (password manager)
3. **Rotate API keys** periodically (at least annually)
4. **Use separate** API keys for different projects
5. **Delete** the .p12 and .p8 files from your Mac after encoding (keep secure backups)

## Cost

- **Apple Developer Account**: $99/year (required)
- **Notarization**: Free (included with developer account)
- **GoReleaser notarization**: Free (uses open source `quill` tool)

## References

- [GoReleaser Notarization Docs](https://goreleaser.com/customization/notarize/)
- [Apple Developer ID Certificates](https://developer.apple.com/support/developer-id/)
- [App Store Connect API Keys](https://developer.apple.com/documentation/appstoreconnectapi/creating_api_keys_for_app_store_connect_api)
- [Notarizing macOS Software](https://developer.apple.com/documentation/security/notarizing_macos_software_before_distribution)
