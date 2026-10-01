---
title: Cheap centralized logs with Loki and S3
slug: loki-s3-logs
date: 2026-08-11T00:00:00Z
tags: Loki, S3, Kubernetes, AWS, observability
status: published
summary: One Helm chart puts Kubernetes pod logs and the CloudWatch logs from RDS, Lambda and ElastiCache into Grafana Loki, with S3 as the store. A year of retention costs a dollar or two a month.
---
On a typical AWS setup the logs are in four places. Pod output is in the cluster, RDS writes to CloudWatch, Lambda writes to CloudWatch under a different group, and whatever runs on plain EC2 is in files on the box. Every managed logging product will happily gather all of that for you and charge per gigabyte ingested, which adds up fast. I run Grafana Loki with S3 as the backend instead. It is one Helm chart, one Grafana to search in, and the storage bill for a year of logs is a couple of dollars a month.

The reason it is cheap is that Loki does not index the log text. It indexes a small set of labels per stream and stores the compressed log lines as chunks in S3. A query scans the chunks that match the labels, so there is no per-query charge like Athena and no big index to pay for.

## The shape of it

```
k8s pods      → Fluent Bit DaemonSet ─────────────┐
                                                   ↓
RDS           → CloudWatch ┐                     Loki → S3
ElastiCache   → CloudWatch ┼→ Fluent Bit (CW in) ─↑      ↓
Lambda        → CloudWatch ┘                          Grafana
```

Fluent Bit runs on every node and tails the container logs. The part that is easy to miss is that Fluent Bit also has a `cloudwatch_logs` input, so the same agent can pull RDS, ElastiCache and Lambda logs out of CloudWatch and send them to Loki alongside the pod logs. There is no second pipeline.

```ini
# k8s pod logs
[INPUT]
    Name    tail
    Path    /var/log/containers/*.log
    Parser  json

# RDS logs from CloudWatch
[INPUT]
    Name           cloudwatch_logs
    Log_Group_Name /aws/rds/cluster/your-cluster/postgresql
    AWS_Region     us-east-1

# buffer to disk so a Loki restart does not drop logs
[OUTPUT]
    Name                     loki
    Match                    *
    Host                     loki.monitoring.svc.cluster.local
    storage.total_limit_size 1G
```

## Labels

Loki stays fast as long as the labels are low cardinality. Service, namespace, level and environment are fine. A user id or a request id as a label creates a new stream for every value and the index grows until queries crawl. Keep those in the log body and pull them out at query time:

```logql
# all errors across services
{job="aws-logs"} |= "ERROR"

# parse JSON and filter on a field
{service="auth"} | json | response_code=500

# error rate by level
sum by (level) (count_over_time({service="auth"} | json [5m]))
```

## What it costs

Loki compresses raw logs somewhere between 10 and 20 times, so 10 GB a day of logs lands as roughly 0.5 to 1 GB a day in S3. Lifecycle rules then move it to cheaper tiers before deleting it:

```
0–30 days   → S3 Standard          $0.023/GB
30–90 days  → S3 Infrequent Access $0.0125/GB
90–365 days → S3 Glacier Instant   $0.004/GB
365+ days   → deleted
```

At 10 GB a day that is about $1 to $2 a month for a full year of history. Glacier Instant Retrieval returns objects in milliseconds, so old logs are queried the same way as new ones. You widen the time range in Grafana and wait a few seconds longer while the chunks come off S3.

```hcl
resource "aws_s3_bucket_lifecycle_configuration" "loki_logs" {
  bucket = aws_s3_bucket.loki.id
  rule {
    id     = "loki-log-retention"
    status = "Enabled"
    transition { days = 30  storage_class = "STANDARD_IA" }
    transition { days = 90  storage_class = "GLACIER_IR" }
    expiration { days = 366 }   # one day after Loki's own retention
  }
}
```

## The CloudWatch double bill

RDS, Lambda and ElastiCache write to CloudWatch whether you want them to or not. Once Fluent Bit is copying those logs into Loki you are paying CloudWatch to keep a second copy. Set those log groups to one day of retention. That is long enough for Fluent Bit to collect them and short enough that CloudWatch stops being a line item.

```hcl
resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/your-function"
  retention_in_days = 1
}
```

## Access

Block public access on the bucket and turn on server-side encryption. Loki itself has no authentication, so anyone who can reach it inside the cluster can read every log. Access control lives in Grafana, with its org and team permissions, and Loki should not be reachable from anywhere else.

## Installing it

```bash
helm repo add grafana https://grafana.github.io/helm-charts
helm install loki-stack grafana/loki-stack \
  --namespace monitoring --create-namespace \
  --set loki.storage.type=s3 \
  --set loki.storage.s3.bucketnames=your-log-bucket \
  --set grafana.enabled=true --set fluent-bit.enabled=true
```

```yaml
loki:
  storage: { type: s3, s3: { bucketnames: your-log-bucket, region: us-east-1 } }
  compactor: { retention_enabled: true, delete_request_store: s3 }
  limits_config: { retention_period: 8760h }   # 365 days
```

On a small to medium cluster the whole thing, Fluent Bit on each node plus one Loki and one Grafana, uses about 1.5 CPU and 3 GB of memory. That and the S3 bill is the entire cost of logging.
