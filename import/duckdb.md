---
title: Querying exported CSVs with DuckDB
slug: duckdb
date: 2026-08-12T00:00:00Z
tags: DuckDB, SQL, data, analytics
status: published
summary: DuckDB is an in-process SQL database like SQLite, but columnar and built for analytics. I use it to run SQL straight over CSV exports on my laptop without loading anything first.
---
Most of the data I need to look at arrives as a CSV. A database dump, the export button in some dashboard, a report someone sent me. For years the options were a spreadsheet or a throwaway pandas script. Now I use DuckDB.

DuckDB is an in-process SQL database in the same spirit as SQLite. There is no server, no port and no daemon. It is either a library you import or a single CLI binary. The difference is that SQLite stores rows and is built for transactions, while DuckDB stores columns and is built for scanning, grouping and joining a lot of rows at once. Same shape of tool, opposite kind of workload.

## You query the file directly

This is the feature that made it stick. There is no import step. The file name goes where a table name would go:

```sql
SELECT count(*), avg(amount)
FROM 'transactions.csv';

SELECT category, sum(amount) AS total
FROM 'transactions.csv'
GROUP BY category
ORDER BY total DESC;
```

It reads CSV, Parquet and JSON, works out the schema itself, and because it is columnar a query that touches two of forty columns only reads those two.

From the shell it is a one-liner:

```bash
duckdb -c "SELECT status, count(*) FROM 'export.csv' GROUP BY status"
```

A folder of exports can be treated as one table with `FROM 'exports/*.csv'`. When the type sniffing guesses wrong, which happens with dates and with columns that are mostly numbers plus one stray string, `read_csv` takes explicit options. If I want to keep a cleaned up result, `COPY (...) TO 'clean.parquet'` writes it back out.

In Python it hands back a DataFrame without copying:

```python
import duckdb
df = duckdb.sql("SELECT * FROM 'export.csv' WHERE amount > 1000").df()
```

It can also read Parquet over HTTP or from S3 with the httpfs extension. I have not needed that. My data is already on the laptop and nothing gets uploaded anywhere, which is part of why I like this workflow.

## Where it does not fit

It is not a server database. There is no network protocol and only one process can write at a time, so it is the wrong choice for an application with many concurrent writers. That is still SQLite or Postgres. DuckDB is for asking questions about data you already have.

It is MIT licensed, comes out of the CWI database research group, and ships as a CLI binary plus libraries for Python, R, Go and Node. Installing it is downloading one file. If you like SQLite for how little it asks of you, DuckDB asks about the same.
