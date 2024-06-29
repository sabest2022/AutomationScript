# EULA and Document Management System
Overview
This project is designed to handle the management of EULA (End User License Agreement) and other document types. It provides a robust mechanism for uploading, tracking, and managing document versions using Google Cloud Storage (GCS) and a backend server to handle webhooks and authentication.

## Key Features
Webhook Listener
A server listens for webhooks from Bitbucket, specifically handling events such as commits and merges.
Repository Handling
Downloads the repository at a specific commit.
Lists changed files.
Processes each file for upload.
File Upload to GCS
Files are uploaded to Google Cloud Storage using pre-signed URLs.
The current logic initiates an upload, performs the upload via HTTP PUT, and completes the upload process.
Database Integration
Metadata about the files, such as name, path, version, and type, are stored in a database to track and manage document versions.
Authentication
Uses GraphQL for authentication to ensure secure operations.
Versioning
Each document is tracked with a version number to manage updates and ensure users have access to the correct versions of documents they have accepted.
Current Flow
Webhook Trigger

A webhook event triggers the process when a commit or merge occurs in the Bitbucket repository.
Download Repository

The repository is downloaded at the specific commit to a temporary directory.
List and Process Changed Files

The changed files are listed, and each file is processed for upload.
Upload Process

The file content is read, encoded, and uploaded to GCS using a pre-signed URL obtained via a GraphQL mutation.
Database Update

Metadata about the uploaded file is saved to the database for tracking.
Summary
This project provides a robust mechanism for managing EULA and other documents, ensuring they are securely uploaded, versioned, and tracked. The addition of GCS versioning will enhance the system by automatically managing versions of documents, simplifying the handling of updates and deletions. The integration with CI/CD pipelines will streamline the deployment and execution process, making the system more efficient and maintainable.

