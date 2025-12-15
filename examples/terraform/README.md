# Terraform S3 Bucket Setup

Creates an S3 bucket with lifecycle policies: automatically moves old backups to Glacier and deletes them after retention period.

## Quick Start

```bash
cd examples/terraform

cat > terraform.tfvars <<EOF
region          = "us-west-2"
bucket_name     = "my-backups-$(date +%s)"
glacier_period  = 7
deletion_period = 30
EOF

terraform init
terraform apply

export S3_BUCKET=$(terraform output -raw bucket_name)
export AWS_REGION=us-west-2
s3-backup
```

## Variables

| Variable          | Default     | Description                      |
| ----------------- | ----------- | -------------------------------- |
| `region`          | `us-east-1` | AWS region                       |
| `bucket_name`     | (required)  | S3 bucket name (globally unique) |
| `glacier_period`  | `7`         | Days before moving to Glacier    |
| `deletion_period` | `30`        | Days before deletion             |

## Lifecycle Timeline

Day 0-6: Standard S3 (~$0.023/GB/month, instant access)  
Day 7-29: Glacier (~$0.004/GB/month, 3-5 hour retrieval)  
Day 30+: Deleted

## Cleanup

```bash
terraform destroy  # ⚠️ Deletes all backups
```
