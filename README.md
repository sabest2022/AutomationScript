# EULA and Document Management System

## Overview

This project handles the management of EULA (End User License Agreement) and other document types. It provides a robust mechanism for uploading, tracking, and managing document versions using Google Cloud Storage (GCS) and a backend server to handle webhooks and authentication.

## Key Features

### 1. Webhook Listener

- A server listens for webhooks from Bitbucket, specifically handling events such as commits and merges.

### 2. Repository Handling

- Downloads the repository at a specific commit.
- Lists changed files.
- Processes each file for upload.

### 3. File Upload to GCS

- Files are uploaded to Google Cloud Storage using pre-signed URLs.
- The current logic initiates an upload, performs the upload via HTTP PUT, and completes the upload process.

### 4. Database Integration

- Metadata about the files, such as name, path, version, and type, are stored in a database to track and manage document versions.

### 5. Authentication

- Uses GraphQL for authentication to ensure secure operations.

### 6. Versioning

- Each document is tracked with a version number to manage updates and maintain a history of changes.
