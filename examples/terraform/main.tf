provider "aws" {
  region = var.region
  default_tags {
    tags = {
      CreatedBy = "Terraform"
    }
  }
}

resource "aws_s3_bucket" "bucket" {
  bucket = var.bucket_name
}

resource "aws_s3_bucket_lifecycle_configuration" "backup_lifecycle" {
  bucket = aws_s3_bucket.bucket.id

  rule {
    id     = "MoveToGlacier"
    status = "Enabled"

    transition {
      days          = var.glacier_period
      storage_class = "GLACIER"
    }
  }

  rule {
    id     = "DeleteOldBackups"
    status = "Enabled"

    expiration {
      days = var.deletion_period
    }
  }
}