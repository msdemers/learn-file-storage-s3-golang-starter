# Tubely App Backend Architecture

A description of how to configure the AWS resources needed to run this demo project.

## Overview

This project serves static assets from Amazon S3 through Amazon CloudFront. The final setup uses:

- A **private S3 bucket** as the origin
- A **CloudFront distribution** in front of that bucket
- CloudFront access to the bucket via an origin access configuration
- Public users access content through **CloudFront**, not directly through S3

## Prerequisites

Before setting this up, make sure you have:

- An AWS account
- Permission to create:
  - S3 buckets
  - CloudFront distributions
  - IAM users, roles, and policies
- AWS CLI installed and configured (optional, but helpful)

## Architecture

    User -> CloudFront -> Private S3 Bucket

## AWS Setup

### 1. Create the S3 bucket

Create an S3 bucket for the project assets.

Suggested settings:

- Use a globally unique bucket name
- Choose the AWS region appropriate for your project
- Keep **Block all public access** enabled
- Do **not** make the bucket public

Example bucket purpose:

- Store static site files
- Store uploaded assets
- Act as the CloudFront origin

### 2. Upload project files

Upload the files needed by the project into the S3 bucket.

Depending on your project, that may include:

- `index.html`
- CSS files
- JavaScript files
- images or other static assets

### 3. Keep the bucket private

This project is designed so that S3 is **not** publicly accessible.

Important:

- Do not add a public bucket policy
- Do not disable block-public-access settings unless your project specifically requires it
- Access should flow through CloudFront

### 4. Create the CloudFront distribution

Create a CloudFront distribution with the S3 bucket as the origin.

Recommended configuration:

- **Origin**: the S3 bucket
- **Viewer protocol policy**: Redirect HTTP to HTTPS
- **Allowed methods**: `GET, HEAD`
- **Default root object**: `index.html` (if serving a static site)
- **Caching**: Use the default policy unless your app requires something else

### 5. Restrict S3 access to CloudFront

Configure CloudFront so it can access the private S3 bucket.

Use the AWS-recommended origin access mechanism for private S3 origins. Then attach a bucket policy that allows reads from CloudFront and denies direct public access.

In plain terms:

- CloudFront is allowed to fetch objects from S3
- End users are not allowed to fetch directly from S3

### 6. Update the bucket policy

Add a bucket policy that grants read access only to the CloudFront distribution.

This policy should:

- Allow `s3:GetObject`
- Apply to the bucket objects
- Be scoped to CloudFront access only

Do not use a wildcard public principal for a private production setup.

### 7. Test the setup

After CloudFront finishes deploying:

- Open the CloudFront distribution domain name in a browser
- Confirm the app or static files load correctly
- Confirm the S3 object URL is not publicly accessible

## IAM Notes

During development, IAM resources may have been created for learning purposes, such as:

- custom IAM policies
- an IAM role
- an IAM user group
- an IAM user
- access keys

For a clean rebuild:

- Prefer least-privilege permissions
- Avoid long-lived access keys when possible
- Remove unused IAM users and keys after testing

## Local Development

If this repo includes local code to generate or upload files, document that here.

Example:

    # install dependencies
    # run the app
    # build static files
    # upload artifacts

## Deployment Notes

Typical deployment flow:

1. Build the project locally
2. Upload generated static files to the private S3 bucket
3. Serve them through CloudFront
4. If content is cached, create a CloudFront invalidation when needed

## Cleanup

To avoid charges when not using this project:

- Delete or disable the CloudFront distribution
- Delete the S3 bucket contents
- Delete the S3 bucket
- Remove unused IAM policies, roles, users, groups, and access keys

## Security Considerations

- Keep the S3 bucket private
- Serve content through CloudFront
- Use least-privilege IAM permissions
- Rotate or delete access keys that are no longer needed
- Avoid committing AWS credentials to the repository

## Troubleshooting

### CloudFront returns access denied

Possible causes:

- Bucket policy does not allow CloudFront to read objects
- CloudFront origin access is not configured correctly
- Object key/path is wrong

### S3 files work directly but CloudFront does not

Possible causes:

- Distribution is still deploying
- Default root object is missing
- Cached errors need invalidation

### CloudFront works but updates do not appear

Possible causes:

- Cached content is still being served
- A CloudFront invalidation is needed

## License

[TODO]