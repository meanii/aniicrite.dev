---
title: Cheap centralized logs with Loki and S3
slug: loki-s3-logs
date: 2026-08-11T00:00:00Z
tags: Loki, S3, Kubernetes, AWS, observability
status: published
summary: Shipping k8s and AWS service logs into Grafana Loki with S3 as the backend — one Helm chart, and about $1–2/month for a year of retention.
---
Logs end up scattered — pod stdout in the cluster, RDS logs in CloudWatch, Lambda somewhere else — and every managed logging product wants real money to put them back together. I run Grafana Loki with S3 as the backend instead. One Helm chart, one UI for everything, and a year of retention costs a couple of dollars a month.

## Why this stack

- **S3 is the cheapest place to keep logs**, and Loki compresses hard before writing.
- **Low ops.** One Helm chart, and Loki has no schema to manage.
- **No per-query cost**, unlike querying logs in S3 with Athena.
- **One UI.** k8s and AWS service logs land in the same Grafana.
- **Retention with auto-deletion**, so it doesn't grow forever.

## The shape of it

```
k8s pods      → Fluent Bit DaemonSet ─────────────┐
                                                   ↓
RDS           → CloudWatch ┐                     Loki → S3
ElastiCache   → CloudWatch ┼→ Fluent Bit (CW in) ─↑      ↓
Lambda        → CloudWatch ┘                          Grafana
```

Fluent Bit is the shipper. On every node it tails pod logs, and — the part people miss — it also **pulls AWS managed-service logs out of CloudWatch** with its `cloudwatch_logs` input. So RDS, ElastiCache, and Lambda land in the same place as your pods, without a separate pipeline. Loki groups everything into per-label streams, compresses them into chunks, and flushes to S3. Grafana queries it with LogQL.

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

# don't lose logs if Loki is down
[OUTPUT]
    Name                     loki
    Match                    *
    Host                     loki.monitoring.svc.cluster.local
    storage.total_limit_size 1G
```

## The one rule that keeps it fast

Only use **low-cardinality** values as stream labels — `service`, `namespace`, `level`, `env`. Never label by `userId` or `requestId`; high-cardinality labels blow up Loki's index. Keep those in the log body and pull them out at query time with `| json`:

```logql
# all errors across services
{job="aws-logs"} |= "ERROR"

# parse JSON and filter on a field
{service="auth"} | json | response_code=500

# error rate by level
sum by (level) (count_over_time({service="auth"} | json [5m]))
```

## The cost, which is the whole point

Loki compresses raw logs 10–20×, so 10 GB/day of logs becomes ~0.5–1 GB/day on S3. Then S3 lifecycle rules tier it down before deleting:

```
0–30 days   → S3 Standard          $0.023/GB
30–90 days  → S3 Infrequent Access $0.0125/GB
90–365 days → S3 Glacier Instant   $0.004/GB
365+ days   → deleted
```

A full year of logs at 10 GB/day works out to roughly **$1–2/month**. Glacier Instant retrieves in milliseconds, so old logs query the same way as recent ones — just widen the time range in Grafana (recent logs are instant from cache; months-old logs take a few seconds while they come off S3).

```hcl
resource "aws_s3_bucket_lifecycle_configuration" "loki_logs" {
  bucket = aws_s3_bucket.loki.id
  rule {
    id     = "loki-log-retention"
    status = "Enabled"
    transition { days = 30  storage_class = "STANDARD_IA" }
    transition { days = 90  storage_class = "GLACIER_IR" }
    expiration { days = 366 }   # 1 day after Loki's own retention, as a safety net
  }
}
```

## The gotcha that doubles your bill

RDS, Lambda, and ElastiCache write to **CloudWatch by default**, so once Fluent Bit copies those logs into Loki you're paying for both. Set the CloudWatch log groups to **1-day retention** — long enough for Fluent Bit to pick them up, short enough that CloudWatch isn't a second bill:

```hcl
resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/your-function"
  retention_in_days = 1
}
```

## Locking it down

- Block public access on the bucket and turn on SSE (`AES256`).
- Loki has **no auth of its own** — anyone with cluster access can read every log. Put access control at the Grafana layer with team/org permissions.

## Deploying it

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

For a small-to-medium cluster the whole stack — Fluent Bit on each node, one Loki, one Grafana — sits around 1.5 CPU and ~3 GB of RAM. That's the entire logging bill: a couple of dollars of S3 and some spare cluster capacity.
