---
title: DuckDB — SQLite's analytics cousin
slug: duckdb
date: 2026-08-12T00:00:00Z
tags: DuckDB, SQL, data, analytics
status: published
summary: How I use DuckDB to query exported CSVs locally — an in-process, columnar SQL database for the analytics half of the work.
---
I reach for SQLite constantly, but it's built for transactions — lots of small reads and writes, one row at a time. When I actually need to *analyze* data — scan a few hundred thousand rows, group, aggregate, join — it's the wrong shape. DuckDB is the tool for that half. It's an in-process SQL database like SQLite, but columnar and built for analytics.

"In-process" is the part that matters. There's no server to run, no port, no daemon. It's a library you import, or a single CLI binary, and it runs right on your machine.

## The feature that sold me: query files in place

Most databases make you load data before you can query it. DuckDB doesn't. You point SQL straight at a file:

```sql
-- a CSV, no import, no schema, no CREATE TABLE
SELECT count(*), avg(amount)
FROM 'transactions.csv';

-- group and sort, still straight off the file
SELECT category, sum(amount) AS total
FROM 'transactions.csv'
GROUP BY category
ORDER BY total DESC;
```

It reads CSV, Parquet, and JSON natively, figures out the schema itself, and being columnar, a query that touches 2 of 40 columns reads 2 columns, not all 40.

## How I actually use it

Mostly on CSVs I've exported from somewhere — a database dump, an app's "export" button, a report — sitting in a folder on my laptop. Instead of opening them in a spreadsheet or writing a pandas script, I just run SQL over them locally:

```bash
duckdb -c "SELECT status, count(*) FROM 'export.csv' GROUP BY status"
```

That's the whole workflow. Nothing gets uploaded anywhere, there's no database to load into first, and I already know SQL — so a question like "how many of these, grouped by month, over some threshold" is one query instead of a script.

A few things that make it stick for this:

- **Point it at many files at once.** A folder of exports becomes one table: `FROM 'exports/*.csv'`.
- **Messy CSVs mostly just work.** It sniffs types and delimiters; when it guesses wrong, `read_csv` takes explicit options.
- **Save the cleaned result** back out to Parquet or a new CSV with `COPY (…) TO 'clean.parquet'`, if I want to keep it.

If I'm already in Python, the same thing works there and hands back a DataFrame with no copy:

```python
import duckdb
df = duckdb.sql("SELECT * FROM 'export.csv' WHERE amount > 1000").df()
```

(It can also read Parquet over HTTP or S3 with the `httpfs` extension — I just don't need that; my data's already local.)

## DuckDB vs SQLite

They look similar and get used together, but they're opposites by design:

- **SQLite** is a row store built for OLTP — transactions, point lookups, many small writes. It runs your app's state.
- **DuckDB** is a column store built for OLAP — scans, aggregates, and joins over lots of rows at once. It answers questions about your data.

Same "embedded, single file, no server" spirit; different engine for a different job.

## When not to use it

It's not a server database. There's no network protocol and it's single-writer, so it's the wrong pick for many clients doing concurrent writes — that's SQLite or Postgres territory. DuckDB is for the read-heavy, analytical side.

It's MIT-licensed, from the CWI research group, and ships as a CLI binary plus libraries for Python, R, Go, Node, and more — a download, not an install-and-configure afternoon. If you like SQLite for how little it asks of you, DuckDB is the same deal for the analytics half.
