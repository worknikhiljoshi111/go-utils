# Package DB

Package db is a utility for connecting to and interacting with our AWS RDS instance. It will use the correct read or write proxy automatically, based on usage, while allowing for explicit use of only read or write.

We're using the [pgx](https://github.com/jackc/pgx) library to create the connection pools, and because it maps Postgres data types to built-in Go types.

## Environment Setup

The environment variable `AWS_REGION` needs to be set in order for the database credentials secret to be retrieved. For ease of use during local development, create a .env file in project root with this value set, as shown in [.env.sample](../.env.sample).

## Examples

[Query](/db/examples/query/query.go)

[QueryRow](/db/examples/queryRow/queryRow.go)

[Exec](/db/examples/exec/exec.go)

[ExecFunc](/db/examples/execFunc/execFunc.go)

These examples show basic usage of how to work with the database connector. They are intentionally fairly verbose and not optimized, so that the process is shown step-by-step without any further abstractions. The examples can be run from either the project root, or the directory that the individual example is in.

Ex: 
```sh
# from project root
go run db/examples/query/query.go
```

```sh
# from db/examples/query
AWS_REGION=us-west-2 go run query.go
```
