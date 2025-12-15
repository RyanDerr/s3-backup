variable "region" {
  type        = string
  description = "The AWS region to deploy resources in."
  default     = "us-east-1"
}

variable "bucket_name" {
  type        = string
  description = "The name of the S3 bucket to create."
}

variable "glacier_period" {
  type        = number
  description = "Number of days after which to move objects to Glacier storage class."
  default     = 7

  validation {
    condition     = var.glacier_period > 0
    error_message = "The move to Glacier period must be greater than 0."
  }
}

variable "deletion_period" {
  type        = number
  description = "Number of days after which to expire objects."
  default     = 30

  validation {
    condition     = var.deletion_period > 0 && var.deletion_period > var.glacier_period
    error_message = "The expiration period must be greater than 0 and greater than the move to Glacier period."
  }

}