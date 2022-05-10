# LearningBay24-backend

This is the backend of the LearningBay24 E-Learning platform.

## Setup

- copy the `example.config.toml` to `config.toml` and change the necessary values
- (optional) configure the `sqlboiler.toml`
- Install `MySQL` or `MariaDB`
- create a database with the name `learningbay24`
- configure the user/password for your database user in `dbconfig.yml`
- apply migrations with `sql-migrate up`

## Compiling

Install the dependencies:

- go (1.18 or higher)

Then compile the backend:

```go
go build
```

## Usage

Run the backend:

```go
./backend
```
