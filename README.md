# s3-backup

A simple tool to backup your local directories to AWS S3.

[![Go Version](https://img.shields.io/badge/go-1.25.5-blue)](go.mod)
[![License](https://img.shields.io/github/license/RyanDerr/s3-backup)](LICENSE)
[![Release](https://img.shields.io/github/v/release/RyanDerr/s3-backup)](https://github.com/RyanDerr/s3-backup/releases)

## What does this do?

This tool takes your files and uploads them to S3. You can run it once manually, or set it up to run on a schedule using cron.

When you backup your files, they get organized in S3 like this:

```
s3://your-bucket/2025-12-15T14-30-00/documents/report.pdf
```

The timestamp makes it easy to keep track of when each backup happened.

## Features

- Run backups on demand or on a schedule (uses cron syntax)
- Backup multiple directories at once
- Optionally include subdirectories
- All uploads happen in parallel for speed

## Installation

### Docker

Works on any platform that runs Docker (Linux amd64/arm64, including Ubuntu):

```bash
docker pull ghcr.io/ryanderr/s3-backup:latest
```

The Docker image is based on Alpine Linux, but the static binary inside will run on Ubuntu, Debian, or any Linux distribution.

### Download the binary

Grab the latest version for your system from [GitHub Releases](https://github.com/RyanDerr/s3-backup/releases):

```bash
# Linux (also works on Ubuntu)
curl -LO https://github.com/RyanDerr/s3-backup/releases/latest/download/s3-backup-linux-amd64
chmod +x s3-backup-linux-amd64
sudo mv s3-backup-linux-amd64 /usr/local/bin/s3-backup

# macOS (Apple Silicon)
curl -LO https://github.com/RyanDerr/s3-backup/releases/latest/download/s3-backup-darwin-arm64
chmod +x s3-backup-darwin-arm64
sudo mv s3-backup-darwin-arm64 /usr/local/bin/s3-backup
```

### Build it yourself

```bash
git clone https://github.com/RyanDerr/s3-backup.git
cd s3-backup
make build
sudo make install
```

## How to use it

### Quick example

```bash
# Tell it where to backup and which S3 bucket to use
export BACKUP_DIRS=/path/to/backup
export AWS_REGION=us-west-2
export S3_BUCKET=my-backup-bucket

# Run it once
s3-backup

# Or run it on a schedule (every day at 2 AM)
export BACKUP_CRON_SCHEDULE="0 2 * * *"
s3-backup  # This keeps running in the background
```

### Using Docker

**One-time backup:**

```bash
docker run --rm \
  -e BACKUP_DIRS=/data \
  -e AWS_REGION=us-west-2 \
  -e S3_BUCKET=my-bucket \
  -e AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} \
  -e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
  -v /path/to/backup:/data:ro \
  ghcr.io/ryanderr/s3-backup:latest
```

**Keep it running with Docker Compose:**

```yaml
services:
  s3-backup:
    image: ghcr.io/ryanderr/s3-backup:latest
    environment:
      - BACKUP_DIRS=/data/documents,/data/photos
      - BACKUP_RECURSIVE=true
      - BACKUP_CRON_SCHEDULE=0 2 * * * # Daily at 2 AM
      - AWS_REGION=us-west-2
      - S3_BUCKET=my-backup-bucket
      - AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}
      - AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}
    volumes:
      - /path/to/documents:/data/documents:ro
      - /path/to/photos:/data/photos:ro
    restart: unless-stopped
```

## Configuration

### Environment variables

| Variable               | Required? | Default       | What it does                                                |
| ---------------------- | --------- | ------------- | ----------------------------------------------------------- |
| `BACKUP_DIRS`          | Yes       | -             | Which directories to backup (separate multiple with commas) |
| `AWS_REGION`           | Yes       | -             | Your AWS region like `us-west-2`                            |
| `S3_BUCKET`            | Yes       | -             | Name of your S3 bucket                                      |
| `BACKUP_RECURSIVE`     | No        | `false`       | Set to `true` to include subdirectories                     |
| `BACKUP_CRON_SCHEDULE` | No        | `0 0 */3 * *` | When to run (default: every 3 days at midnight)             |

### Using a config file

You can also put everything in a YAML file:

```yaml
backup_dirs:
  - /home/user/documents
  - /home/user/photos
recursive: true
cron_schedule: "0 2 * * *"
aws_region: us-west-2
s3_bucket: my-backup-bucket
```

Then run:

```bash
export S3_BACKUP_CONFIG_FILE=config.yaml
s3-backup
```

Check out the [examples/](examples/) folder for more ways to configure it.

## Where to find it

**Docker images:** `ghcr.io/ryanderr/s3-backup`

- `latest` - most recent version
- `v1.0.0` - specific version tags
- Works on `linux/amd64` and `linux/arm64`

**Binaries:** [GitHub Releases](https://github.com/RyanDerr/s3-backup/releases)

- Linux (amd64, arm64) - statically compiled, works on Ubuntu, Debian, Alpine, etc.
- macOS (amd64, arm64)
- Windows (amd64)

## Help out

If you find this useful:

- ⭐ [Star the repo](https://github.com/RyanDerr/s3-backup)
- 🐛 [Report bugs](https://github.com/RyanDerr/s3-backup/issues)
- 💡 [Request features](https://github.com/RyanDerr/s3-backup/issues/new)
- 🔀 [Send pull requests](https://github.com/RyanDerr/s3-backup/pulls)

## For developers

More documentation:

- [DOCKER.md](DOCKER.md) - Docker setup details
- [RELEASE.md](RELEASE.md) - How releases work
- [examples/README.md](examples/README.md) - More config examples

Useful commands:

```bash
make test           # Run all tests
make test-coverage  # See what's covered by tests
make test-fuzz      # Run fuzz tests
make build          # Build the binary
make docker-build   # Build Docker image
```

## License

MIT - see [LICENSE](LICENSE)
