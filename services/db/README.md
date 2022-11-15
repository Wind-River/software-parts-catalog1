# db

The db container is based on the postgres image.
To handle schema migrations, we use goose, which is built as the first Docker build stage using the golang image.

Additionally restoring from database dumps is supported by placing them into either docker-entrypoint-initdb.d/tkdb/ or docker-entrypoint-initdb.d/blob/ for their respective databases.

## goose
[Goose](https://github.com/pressly/goose) is a tool we use to migrate, or initialize, a database to our current schema version.

Typically we build a binary that given a database and a directory of sql migrations, will upgrade the database to the most recent schema.
We can additionally include golang migrations in that directory, which are built into the binary, though this is not usually required.

Migrations can be created using `goose create $migration_name` for a golang migration or `goose create $migration_name sql` for an sql migration.
They should be stored in goose/migrations/tkdb or goose/migrations/blob.