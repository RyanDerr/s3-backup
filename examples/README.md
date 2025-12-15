# S3 Backup Examples

This directory contains examples demonstrating how to configure and use the S3 backup tool.

## Table of Contents

- [Configuration Methods](#configuration-methods)
- [S3 Key Structure](#s3-key-structure)
- [Examples](#examples)

## Configuration Methods

### 1. YAML Configuration

Create a `config.yaml` file (see [config.yaml](./config.yaml) for a complete example):

```yaml
backup_dirs:
  - /Users/username/Documents
  - /Users/username/Photos
recursive: true
cron_schedule: "0 0 */3 * *"
aws_region: us-west-2
s3_bucket: my-backup-bucket
```

Set the path via environment variable:

```bash
export S3_BACKUP_CONFIG_FILE=/path/to/config.yaml
./s3-backup
```

### 2. Environment Variables

Use environment variables directly (see [.env.example](./.env.example)):

```bash
export BACKUP_DIRS="/Users/username/Documents,/Users/username/Photos"
export BACKUP_RECURSIVE=true
export BACKUP_CRON_SCHEDULE="0 0 */3 * *"
export AWS_REGION=us-west-2
export S3_BUCKET=my-backup-bucket
./s3-backup
```

### 3. Mixed Configuration

Environment variables override YAML settings:

```bash
# Use YAML for most settings
export S3_BACKUP_CONFIG_FILE=/path/to/config.yaml
# Override specific settings
export BACKUP_RECURSIVE=false
export AWS_REGION=us-east-1
./s3-backup
```

## S3 Key Structure

Files are uploaded to S3 with the following key structure:

```
s3://{bucket}/{timestamp}/{directory-name}/{relative-path}
```

Where:

- `{bucket}`: Your S3 bucket name
- `{timestamp}`: Backup timestamp in format `YYYY-MM-DDTHH-MM-SS`
- `{directory-name}`: Base name of the backup directory
- `{relative-path}`: Relative path from the backup directory

## Examples

### Example 1: Non-Recursive Backup

**Configuration:**

```yaml
backup_dirs:
  - /Users/alice/Documents
recursive: false
```

**Local Directory Structure:**

```
/Users/alice/Documents/
├── report.pdf
├── notes.txt
├── project/
│   └── code.py
└── images/
    └── photo.jpg
```

**Result in S3:**

```
s3://my-backup-bucket/
└── 2025-12-15T14-30-00/
    └── Documents/
        ├── report.pdf
        └── notes.txt
```

Only files in the top-level directory are backed up. Subdirectories (`project/`, `images/`) are skipped.

---

### Example 2: Recursive Backup (Single Directory)

**Configuration:**

```yaml
backup_dirs:
  - /Users/alice/Documents
recursive: true
```

**Local Directory Structure:**

```
/Users/alice/Documents/
├── report.pdf
├── notes.txt
├── project/
│   ├── code.py
│   └── data/
│       └── config.json
└── images/
    └── photo.jpg
```

**Result in S3:**

```
s3://my-backup-bucket/
└── 2025-12-15T14-30-00/
    └── Documents/
        ├── report.pdf
        ├── notes.txt
        ├── project/
        │   ├── code.py
        │   └── data/
        │       └── config.json
        └── images/
            └── photo.jpg
```

All files and subdirectories are backed up, preserving the directory structure.

---

### Example 3: Multiple Directories (Recursive)

**Configuration:**

```yaml
backup_dirs:
  - /Users/alice/Documents
  - /Users/alice/Photos
  - /var/log/myapp
recursive: true
```

**Local Directory Structure:**

```
/Users/alice/Documents/
├── report.pdf
└── project/
    └── code.py

/Users/alice/Photos/
├── vacation.jpg
└── 2025/
    └── summer.jpg

/var/log/myapp/
├── app.log
└── errors/
    └── error.log
```

**Result in S3:**

```
s3://my-backup-bucket/
└── 2025-12-15T14-30-00/
    ├── Documents/
    │   ├── report.pdf
    │   └── project/
    │       └── code.py
    ├── Photos/
    │   ├── vacation.jpg
    │   └── 2025/
    │       └── summer.jpg
    └── myapp/
        ├── app.log
        └── errors/
            └── error.log
```

Each backup directory maintains its own prefix based on the directory's base name.

---

### Example 4: Multiple Directories (Non-Recursive)

**Configuration:**

```yaml
backup_dirs:
  - /Users/alice/Documents
  - /Users/alice/Photos
recursive: false
```

**Local Directory Structure:**

```
/Users/alice/Documents/
├── report.pdf
├── notes.txt
└── project/
    └── code.py

/Users/alice/Photos/
├── vacation.jpg
└── 2025/
    └── summer.jpg
```

**Result in S3:**

```
s3://my-backup-bucket/
└── 2025-12-15T14-30-00/
    ├── Documents/
    │   ├── report.pdf
    │   └── notes.txt
    └── Photos/
        └── vacation.jpg
```

Only top-level files from each directory are backed up.

---

### Example 5: Scheduled Backups with Cron

**Configuration:**

```yaml
backup_dirs:
  - /Users/alice/Documents
recursive: true
cron_schedule: "0 2 * * 0" # Every Sunday at 2 AM
aws_region: us-west-2
s3_bucket: weekly-backups
```

**Result:**

The backup tool will run continuously and execute backups every Sunday at 2 AM. Each backup creates a new timestamp prefix:

```
s3://weekly-backups/
├── 2025-12-07T02-00-00/
│   └── Documents/
│       └── ...
├── 2025-12-14T02-00-00/
│   └── Documents/
│       └── ...
└── 2025-12-21T02-00-00/
    └── Documents/
        └── ...
```

**Common Cron Schedules:**

- `"0 0 * * *"` - Daily at midnight
- `"0 0 */3 * *"` - Every 3 days at midnight (default)
- `"0 2 * * 0"` - Weekly on Sunday at 2 AM
- `"0 3 1 * *"` - Monthly on the 1st at 3 AM
- `"*/30 * * * *"` - Every 30 minutes

---

## Running the Tool

### One-time Backup (No Cron)

If you don't set a cron schedule and use the tool for a one-time backup:

```bash
# Using config file
export S3_BACKUP_CONFIG_FILE=config.yaml
./s3-backup

# Using environment variables
export BACKUP_DIRS=/path/to/backup
export AWS_REGION=us-west-2
export S3_BUCKET=my-bucket
./s3-backup
```

### Scheduled Backups (With Cron)

The tool will run continuously, executing backups on the specified schedule:

```bash
export S3_BACKUP_CONFIG_FILE=config.yaml
./s3-backup
# Tool runs in foreground, executing backups based on cron_schedule
```

To run as a background service:

```bash
nohup ./s3-backup > backup.log 2>&1 &
```

Or use systemd, Docker, or your preferred service manager.

---

## Notes

1. **Timestamp Format**: `YYYY-MM-DDTHH-MM-SS` (e.g., `2025-12-15T14-30-00`)
2. **Base Directory Names**: Only the base name is used (e.g., `/Users/alice/Documents` → `Documents`)
3. **Path Separators**: S3 keys use `/` regardless of OS
4. **Concurrency**: All directories are backed up in parallel
5. **Context Cancellation**: Press Ctrl+C to gracefully stop the backup
6. **Default Schedule**: If not specified, defaults to every 3 days at midnight (`0 0 */3 * *`)
