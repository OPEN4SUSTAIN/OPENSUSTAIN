---
layout: default
title: GitHub App Setup
---

# GitHub App Authentication Setup

For large organizations, GitHub App authentication provides significantly higher rate limits (5,000 requests/hour vs 60/hour for PATs).

## Why Use GitHub App Authentication?

<ul class="feature-list">
  <li><strong>83x higher rate limits</strong> (5,000/hour vs 60/hour)</li>
  <li><strong>Better suited for enterprise organizations</strong></li>
  <li><strong>More secure than sharing PATs</strong></li>
  <li><strong>Centralized app management</strong></li>
</ul>

## Setting up a GitHub App

### 1. Create a GitHub App

1. Go to GitHub Settings → Developer settings → GitHub Apps → New GitHub App
2. Give it a name (e.g., "OpenSustain Scanner")
3. Set Homepage URL to your project URL
4. Uncheck "Webhook" (not needed for this use case)
5. Click "Create GitHub App"

### 2. Configure Permissions

**Repository permissions** (set to Read access):

<ul class="feature-list">
  <li><strong>Contents</strong>: Read</li>
  <li><strong>Issues</strong>: Read</li>
  <li><strong>Pull requests</strong>: Read</li>
</ul>

**Organization permissions** (set to Read access):

<ul class="feature-list">
  <li><strong>Administration</strong>: Read (to list installations)</li>
</ul>

### 3. Generate Private Key

1. In the GitHub App settings, scroll to "Private keys"
2. Click "Generate a private key"
3. Download the `.pem` file and keep it secure
4. Note the **App ID** from the GitHub App settings page

### 4. Install the App

1. Click "Install App" in the GitHub App settings
2. Select the organizations you want to scan
3. Click "Install" on each organization

## Using GitHub App Authentication

```bash
./OpenSustain scan org \
  --org my-github-org \
  --days 90 \
  --app-id 123456 \
  --private-key-path /path/to/private-key.pem
```

## Finding Your App ID

1. Go to GitHub → Settings → Developer settings → GitHub Apps
2. Click on your App name
3. The App ID is displayed at the top of the page (6-digit number)

## Security Notes

<ul class="feature-list">
  <li>Keep your private key file secure and never commit it to version control</li>
  <li>The private key should have restricted file permissions (chmod 600)</li>
  <li>Rotate your private key if it's ever compromised</li>
  <li>Only grant the minimum required permissions</li>
</ul>

## Troubleshooting

**Error: "failed to read private key"**

<ul class="feature-list">
  <li>Ensure the private key file path is correct</li>
  <li>Check file permissions on the private key file</li>
</ul>

**Error: "no installation found for organization"**

<ul class="feature-list">
  <li>Ensure the GitHub App is installed on the target organization</li>
  <li>Check that the organization name is correct</li>
</ul>

**Error: "failed to get installation token"**

<ul class="feature-list">
  <li>Verify the App has the correct permissions</li>
  <li>Check that the App is not rate-limited</li>
  <li>Ensure the private key is valid and not expired</li>
</ul>
