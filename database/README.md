# Database

The database directory contains all schema and queries needed for each database. Each sub-directory inside this directory represents a database.

For example,

```text
|-database
    |- ledger
         |- schema.sql
         |- query.sql
```

This means `ledger` is a database.

## Database Schema

We define database schema in one(1) file, usually in a file named `schema.sql`. The schema is being defined inside one file so its easier for the user to understand the whole schema and its faster for us to apply the schema rather applying all the migration schemas.

Further question might be, then how we apply the database migration and applying schema evolution to our production database? We prefer to use something like [pg-diff](https://github.com/michaelsogos/pg-diff) or [pg-schema-diff](https://github.com/stripe/pg-schema-diff) to produce a `diff` of our database and apply the changes accordingly. The reason behind this is because applying a schema migration on a live database need to be done carefully. Let's keep this topic for another day.

## SQLC

We use [sqlc](https://sqlc.dev/) to automatically generate the queries into our `Go` code. All queries need to be defined inside a `sql` file named `query.sql` inside the database directory. While its not perfect and it has issues, it's the near-perfect solution for us as we want to ue `raw` query as much as possible.

Why query generator rather than using an ORM? In my opinion, to have a proper conversation with a database, you need to talk with the language the database understand. It's worth it to learn SQL and doing some `raw` SQL directly to the database as there might be some features that not supported by the ORMs. For example, CTE.

## SQLFluff

Because the usage of `.sql` whenever we can, we can also use [sqlfluff](https://sqlfluff.com/) to lint our queries.

The `sqlfluff` linter is used so we can have a standard across the repository. Producing codes is one thing, and maintaining codes is another thing. Its important to make things consistent across the repository/organization.

## Helper Script

Within this directory we provide a `helper` script to:

1. Creating new directory/database.

   The script will create the directory and copy the `sqlc.yaml` in the `./database` directory and replace `database_name` string.

   ```shell
   $./all.sh templategen [database_name]
   ```

1. Replacing all `sqlc.yaml`.

   Sometimes we want to update the `sqlc.yaml` and apply the changes to all databases.

   ```shell
   $./all.sh templategen
   ```

1. Generating `go` code via `sqlc`.

   As we said earlier, we use `sqlc` to generate the database layer code.

   To generate code for a specific database use:

   ```shell
   $./all.bash generate [database_name]
   ```

   To generate code for all databases use:

   ```shell
   $./all.bash generate
   ```
