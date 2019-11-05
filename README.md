# migrations-renderer

Solves the following problem: get the latest database schema overview without having to connect to any running database in any particular environment.

This utility takes given [Flyway](https://github.com/flyway/flyway) migration files and outputs the latest database schema to stdout as it is after applying all of the migrations.

## Usage

```bash
$ sh migrations-renderer.sh /path/to/migrations
```
