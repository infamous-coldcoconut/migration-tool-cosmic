1. create a env file (.env) in root path
2. add DATABASE_URL for your database, MIGRATIONS_DIR for your change tables folder, MIGRATIONLOCKID for pg_advisory_unlock (can be random id)
   example:

DATABASE_URL=postgres://postgres:admin@localhost:5432/cosmic?sslmode=disable
MIGRATIONS_DIR=./migrations
MIGRATIONLOCKID=42424242
MIGRATION_AUTHOR=admin
