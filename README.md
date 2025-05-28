# Migrator
Package to run SQL migrations from within your Go application. It supports upgrading and downgrading your database.

Supported databases:
* **SQLite** (tested on version 3)
* **PostgreSQL** (tested on 17.4 but should work for older versions as well)

This package only depends on database drivers for testing. When in regular use it only uses the standard package `database/sql`.

## Migrations file
Your migrations are store in a YAML file in the following format:
```yaml
migrations:
    # comment is an optional field with details about your migration step
  - comment: "My first migration"
    # up is the SQL statement executed when upgrading your database. This field is mandatory.
    up: >
      CREATE TABLE test
      (id INTEGER PRIMARY KEY)
    # down is the SQL statement executed when downgrading, usually it reverses the
    # effect of an upgrade. This field is optional.
    down: >
      DROP TABLE test
...
```
A migration must atleast have an `up`-statement to be valid.

## Environment variables
At run-time there are two environment variables that must be set:
* `MIGRATOR_FILE`: migration file written in YAML
* `MIGRATOR_TARGET_VERSION`: version number you want to migrate to

## About versions
Current version is stored in the database. Table storing your version might differ between Migrator implementations. Calling the New-method for a migrator will setup migrations in the given database and return a Migrator ready to run migrations. Your database will now be at version 0, i.e. no migrations have been run.

Setting `MIGRATOR_TARGET_VERSION` to 1 at version 0 and running `Migrate()` will execute the first `up` statement in your YAML file. If executed without errors your database will be at version 1.

If you run your migration again but setting `MIGRATOR_TARGET_VERSION` to 0 it will run the your `down` statement and the database will be at version 0.

## Example
There is also a working example in [tesdata/example](testdata/example).

1. Write your application in, `main.go`:
    ```golang
    package main

    import (
        "database/sql"
        "log"

        _ "github.com/mattn/go-sqlite3"
        "github.com/spagettikod/migrator"
    )

    func main() {
        // Setup your database connection for your application
        db, err := sql.Open("sqlite3", ":memory:")
        if err != nil {
            log.Fatal(err)
        }

        // Create a migrator for your database
        // (it will initialize it self if run for the first time)
        migrator, err := migrator.NewSqliteMigrator(db)
        if err != nil {
            log.Fatal(err)
        }

        // Run the migrations in up/down to the given target version
        _, err = migrator.Migrate()
        if err != nil {
            log.Fatal(err)
        }
    }
    ```
1. Write your migration script, this example is saved in `migrations.yml`:
    ```yaml
    migrations:
    - comment: "Create user table"
        up: >
            CREATE TABLE user
            (id INTEGER PRIMARY KEY) 
        down: >
            DROP TABLE user
    - comment: "Create address table"
        up: >
            CREATE TABLE address
            (id INTEGER PRIMARY KEY) 
        down: >
            DROP TABLE address
    ```
1. Run your application, at startup it will run the given migrations:
    ```bash
    MIGRATOR_FILE=migrations.yml MIGRATOR_TARGET_VERSION=2 go run main.go
    ```
