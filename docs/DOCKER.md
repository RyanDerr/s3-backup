# Docker Guide

How to run s3-backup using Docker.

## Getting the image

```bash
docker pull ghcr.io/ryanderr/s3-backup:latest
```

Works on `linux/amd64` and `linux/arm64`.

## Quick start

Run a one-time backup:

```bash
docker run --rm \
  -e BACKUP_DIRS=/data \
  -e AWS_REGION=us-west-2 \
  -e S3_BUCKET=my-bucket \
  -e AWS_ACCESS_KEY_ID=your-key \
  -e AWS_SECRET_ACCESS_KEY=your-secret \
  -v /path/to/backup:/data:ro \
  ghcr.io/ryanderr/s3-backup:latest
```

### Multiple directories

```bash
docker run --rm \
  -e BACKUP_DIRS=/data/docs,/data/photos \
  -e BACKUP_RECURSIVE=true \
  -e AWS_REGION=us-west-2 \
  -e S3_BUCKET=my-bucket \
  -e AWS_ACCESS_KEY_ID=your-key \
  -e AWS_SECRET_ACCESS_KEY=your-secret \
  -v /path/to/docs:/data/docs:ro \
  -v /path/to/photos:/data/photos:ro \
  ghcr.io/ryanderr/s3-backup:latest
```

### Run on a schedule

Keep it running in the background:

```bash
docker run -d \
  --name s3-backup \
  --restart unless-stopped \
  -e BACKUP_DIRS=/data \
  -e BACKUP_CRON_SCHEDULE="0 2 * * *" \
  -e AWS_REGION=us-west-2 \
  -e S3_BUCKET=my-bucket \
  -e AWS_ACCESS_KEY_ID=your-key \
  -e AWS_SECRET_ACCESS_KEY=your-secret \
  -v /path/to/backup:/data:ro \
  ghcr.io/ryanderr/s3-backup:latest
```

Check if it's running:

```bash
docker ps | grep s3-backup
docker logs -f s3-backup
```

## Using a config file

Create `config.yaml`:

Run with the config:

```bash
docker run --rm \
  -v $(pwd)/config.yaml:/config.yaml:ro \
  -v /path/to/docs:/data/docs:ro \
  -v /path/to/photos:/data/photos:ro \
  -e S3_BACKUP_CONFIG_FILE=/config.yaml \
  -e AWS_ACCESS_KEY_ID=your-key \
  -e AWS_SECRET_ACCESS_KEY=your-secret \
  ghcr.io/ryanderr/s3-backup:latest
```

## Docker Compose

Create `docker-compose.yml`:

```yaml
services:
  s3-backup:
    image: ghcr.io/ryanderr/s3-backup:latest
    restart: unless-stopped
    environment:
      - BACKUP_DIRS=/data/documents,/data/photos
      - BACKUP_RECURSIVE=true
      - BACKUP_CRON_SCHEDULE=0 2 * * *
      - AWS_REGION=us-west-2
      - S3_BUCKET=my-backup-bucket
      - AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}
      - AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}
    volumes:
      - /path/to/documents:/data/documents:ro
      - /path/to/photos:/data/photos:ro
```

Start it:

```bash
docker-compose up -d
docker-compose logs -f
```

### With a config file

```yaml
services:
  s3-backup:
    image: ghcr.io/ryanderr/s3-backup:latest
    restart: unless-stopped
    environment:
      - S3_BACKUP_CONFIG_FILE=/config/config.yaml
      - AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}
      - AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}
    volumes:
      - ./config.yaml:/config/config.yaml:ro
      - /path/to/backup:/data:ro
```

## Environment variables

| Variable                | Required? | Default       | What it does                        |
| ----------------------- | --------- | ------------- | ----------------------------------- |
| `BACKUP_DIRS`           | Yes\*     | -             | Comma-separated list of directories |
| `AWS_REGION`            | Yes\*     | -             | AWS region (e.g., `us-west-2`)      |
| `S3_BUCKET`             | Yes\*     | -             | Your S3 bucket name                 |
| `BACKUP_RECURSIVE`      | No        | `false`       | Include subdirectories              |
| `BACKUP_CRON_SCHEDULE`  | No        | `0 0 */3 * *` | When to run (every 3 days)          |
| `S3_BACKUP_CONFIG_FILE` | No        | -             | Path to config file                 |
| `AWS_ACCESS_KEY_ID`     | Yes       | -             | AWS access key                      |
| `AWS_SECRET_ACCESS_KEY` | Yes       | -             | AWS secret key                      |

\*Not required if you're using a config file

## Using AWS credentials file

Instead of passing keys as environment variables:

```bash
docker run --rm \
  -v ~/.aws:/home/s3backup/.aws:ro \
  -e AWS_PROFILE=default \
  -e BACKUP_DIRS=/data \
  -e AWS_REGION=us-west-2 \
  -e S3_BUCKET=my-bucket \
  -v /path/to/backup:/data:ro \
  ghcr.io/ryanderr/s3-backup:latest
```

## Troubleshooting

**Check if it's running:**

```bash
docker ps | grep s3-backup
```

**View logs:**

```bash
docker logs s3-backup
docker logs -f s3-backup  # follow logs
```

**Permission issues:**

```bash
chmod -R +r /path/to/backup
```

**Test the image:**

```bash
docker run --rm ghcr.io/ryanderr/s3-backup:latest --version
docker run --rm ghcr.io/ryanderr/s3-backup:latest --help
```

## Building it yourself

```bash
docker build -t s3-backup:local .
docker run --rm s3-backup:local --version
```
