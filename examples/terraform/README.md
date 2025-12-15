# Terraform S3 Bucket Setup

Creates an S3 bucket with lifecycle policies for s3-backup: automatically moves old backups to Glacier and deletes them after retention period.

## Quick Start

```bash
cd examples/terraform

# Create config
cat > terraform.tfvars <<EOF
region          = "us-west-2"
bucket_name     = "my-backups-$(date +%s)"
glacier_period  = 7   # Move to Glacier after 7 days
deletion_period = 30  # Delete after 30 days
EOF

terraform init
terraform apply
```

## Variables

| Variable          | Default     | Description                      |
| ----------------- | ----------- | -------------------------------- |
| `region`          | `us-east-1` | AWS region                       |
| `bucket_name`     | (required)  | S3 bucket name (globally unique) |
| `glacier_period`  | `7`         | Days before moving to Glacier    |
| `deletion_period` | `30`        | Days before deletion             |

## Lifecycle Timeline

Default settings (`glacier_period=7`, `deletion_period=30`):

```
Day 0-6:  Standard S3 (fast access, ~$0.023/GB/month)
Day 7-29: Glacier (cheap, ~$0.004/GB/month, 3-5 hour retrieval)
Day 30+:  Deleted
```

## Use with s3-backup

```bash
export S3_BUCKET=$(terraform output -raw bucket_name)
export AWS_REGION=us-west-2
s3-backup
```

## Cleanup

⚠️ Deletes all backups:

```bash
terraform destroy
```

# Get bucket size and storage class breakdown

aws cloudwatch get-metric-statistics \
 --namespace AWS/S3 \
 --metric-name BucketSizeBytes \
 --dimensions Name=BucketName,Value=$(terraform output -raw bucket_name) \
 --start-time $(date -u -d '7 days ago' +%Y-%m-%dT%H:%M:%S) \
 --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
 --period 86400 \
 --statistics Average

````

## Troubleshooting

**Lifecycle policies not applying:**

- Lifecycle transitions happen at midnight UTC
- It may take 24-48 hours for policies to take effect
- Check the bucket's lifecycle configuration is correct

**Access denied when retrieving from Glacier:**

- Glacier objects require a restore request before access
- Restore takes 3-5 hours for standard retrieval

```bash
aws s3api restore-object \
  --bucket your-bucket \
  --key 2025-12-01T00-00-00/Documents/file.pdf \
  --restore-request Days=1
````

**Bucket name already exists:**

- S3 bucket names must be globally unique
- Add a timestamp or random suffix to your bucket name

```hcl
bucket_name = "my-backups-${formatdate("YYYYMMDD", timestamp())}"
```

## Additional Resources

- [Terraform AWS Provider Documentation](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [S3 Lifecycle Configuration Guide](https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lifecycle-mgmt.html)
- [S3 Storage Classes Comparison](https://aws.amazon.com/s3/storage-classes/)
- [S3 Pricing Calculator](https://calculator.aws/#/addService/S3
  ```bash
  terraform apply
  ```

5. **Use the bucket with s3-backup**

   ```bash
   export S3_BUCKET=my-s3-backup-bucket
   export AWS_REGION=us-west-2
   s3-backup
   ```

## Variables

| Variable          | Type   | Required | Default     | Description                                         |
| ----------------- | ------ | -------- | ----------- | --------------------------------------------------- |
| `region`          | string | No       | `us-east-1` | AWS region to create the bucket in                  |
| `bucket_name`     | string | Yes      | -           | Name of the S3 bucket (must be globally unique)     |
| `glacier_period`  | number | No       | `7`         | Days before moving objects to Glacier storage class |
| `deletion_period` | number | No       | `30`        | Days before permanently deleting objects            |

## Lifecycle Policy Details

### Glacier Transition

After `glacier_period` days, objects are automatically moved to the Glacier storage class:

- **Glacier**: Long-term archival storage, much cheaper than standard S3
- **Retrieval**: Takes hours to retrieve data (not instant like standard S3)
- **Use case**: Old backups you rarely need to restore

### Automatic Deletion

After `deletion_period` days, objects are permanently deleted:

- Helps control storage costs
- Ensures compliance with data retention policies
- Must be greater than `glacier_period`

### Example Timeline

With default settings (`glacier_period=7`, `deletion_period=30`):

```
Day 0:  Backup created in Standard S3 storage
Day 7:  Moved to Glacier (cheaper storage)
Day 30: Permanently deleted
```

## Cost Optimization

Adjust the lifecycle periods based on your needs:

**Frequent backups, short retention:**

```hcl
glacier_period  = 3
deletion_period = 14
```

**Long-term archival:**

```hcl
glacier_period  = 30
deletion_period = 365  # Keep for 1 year
```

**No Glacier, just deletion:**

```hcl
# Modify main.tf to comment out the Glacier transition rule
deletion_period = 90
```

## Outputs

After applying, Terraform will output:

- `bucket_name`: The name of the created S3 bucket
- `bucket_arn`: The ARN of the bucket (for IAM policies)

## Cleanup

To destroy the bucket (⚠️ this will delete all backups):

```bash
terraform destroy
```

## Notes

- **Bucket names must be globally unique** across all AWS accounts
- **Versioning**: This example doesn't enable versioning. Add it if needed.
- **Encryption**: Consider adding server-side encryption in production
- **Access policies**: Add bucket policies or IAM roles as needed

## Example with Encryption

To add server-side encryption, modify `main.tf`:

```hcl
resource "aws_s3_bucket_server_side_encryption_configuration" "backup_encryption" {
  bucket = aws_s3_bucket.bucket.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}
```

## Additional Resources

- [Terraform AWS Provider Documentation](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [S3 Lifecycle Configuration](https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lifecycle-mgmt.html)
- [S3 Storage Classes](https://aws.amazon.com/s3/storage-classes/)
