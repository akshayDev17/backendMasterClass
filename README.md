1. Created DB Schema using [dbdiagram.io](dbdiagram.io) -> simple bank.sql \
    <img src="simple bank.png">
2. Searched for PostGreS docker image on [hub.docker.com](https://hub.docker.com/_/postgres)
    1. latest one was used, alpine: smaller-sized image
    2. `docker run --name some-postgres -e POSTGRES_PASSWORD=mysecretpassword -d postgres` (-d: base image)
3. list docker containers running currently: `docker ps`, list images : `docker images`, `-p host_port:container_port` , run command in a container: `docker exec -it` (it: interactive ttyl)
    `docker logs container-name` : see logs of this container
4. Table plus: database management with UI.
    1. Default port for PostGres while using TablePlus is 5432, hence that was chosen while setting up the container.
    2. Create a new connection(s) in Table Plus by choosing the driver followed by the username and password , the host and the port connected.
    3. Open files using Cmd+O, open databases using Cmd+K
5. DB Migration in Go - done to adapt to new business requirements.
    1. [golang migrate](https://pkg.go.dev/github.com/golang-migrate/migrate/v4#section-readme), [BEST PRACTICES FOR WRITING MIGRATION FILES](https://github.com/golang-migrate/migrate/blob/v4.15.2/MIGRATIONS.md) \
        `brew install golang-migrate`
    2. `migrate -version`, `migrate -help` (manual)
    3. store all migration files: `mkdir -p db/migration`
    4. create a migration file: `migrate create -ext sql -dir db/migration -seq init_schema`
        1. `-ext`: extension is sql
        2. migration directory is `-dir` 
        3. `-seq`: sequential version number for this migration file
        4. `init_schema`: migration file name
        5. What is up/down migrations?
            1. <img src="up-down-migration.png" />
            2. update old schema: up script is run, revert changes to an older schema: down script is run.
            3. 1.sql-->2.sql-->3.sql--> UP, DOWN: 3.sql->2.sql->1.sql (3.sql means 3_down.sql)
        6. init_schema_1_up.sql == sql from dbdiagram.io, first down script = drop all tables created in first up script. \
        Note: You need to fill up the contents of all up/down migration sql files.
6. Makefile was created to:
    1. create a docker container based on postgres
    2. create a postgres db inside this container
    3. [**Please go through this resource**](https://makefiletutorial.com).
    4. `.PHONY` is used to make a target, a dependency of a special target called [.PHONY](https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html), so that if , say a file `createdb` exists, make createdb runs regardless.
7. Perform migration in `simple_bank` database(postgres DB inside a docker container)
    1. `migrate -path db/migration -database "postgresql://postgres:mysecretpassword@localhost:5432/simple_bank?sslmode=disable" -verbose up`
        1. `migrate -help` to understand meaning of each flag
        2. notice , even in [migrate's basic usage documentation](https://github.com/golang-migrate/migrate) you will see `-source file://path/to/migrations` but a shorthand for this is `-path db/migration`
            1. `db/migration` is the relative path.
        3. `driver` = database driver, here `postgresql`
            1. [format](http://www.postgresql.org/docs/current/static/libpq-connect.html#LIBPQ-CONNSTRING) of postgresql connection string(url) = `postgresql://[user[:password]@][netloc][:port][/dbname][?param1=value1&...]`(without the box brackets)
            2. `sslmode` is a parameter, which we want to currently disable since `pq: SSL is not enabled on the server`.(connection is established without SSL) \
                [From AWS](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/PostgreSQL.Concepts.General.SSL.html)
                > Using SSL, you can encrypt a PostgreSQL connection between your applications and your PostgreSQL DB instances. \
                For general information about SSL support and PostgreSQL databases, see [SSL support](https://www.postgresql.org/docs/11/libpq-ssl.html) in the PostgreSQL documentation.
        4. `up` = you will have to specify whether you've to migrate up or down.
            1. > the order of dropping tables matter, because `entries` and `transfers` have `account_id` related
                foreign key constraint-dependency on `accounts` table.
                hence while migrating down if `accounts` is dropped first,
                `entries` and `transfers` though existent will not have those constraints
                satisfied, hence will throw an error.
            2. > similar is the case for Up, wherein the table that isn't dependent should be created first, followed by tables dependent on this.
    2. Refresh TablePlus to see schema_migration as one of the tables, with a dirty column
        1. If this is false, no worries, if True, migration failed, fix issues manually and make DB state cleam then before any other migration versions are run.
        2. `migrate_force_v1` is used to force migrate to version 1(in case of dirty migration, force migrate to this version to fix errors).
8. Gorm can run 3-5x slower than normal DB/SQL calls according to some benchmarks on the internet. **FIND THIS !!!!**
9. SQL-C generates the code first, which is then integrated, hence any errors are informed already whilst generation of these files.
    <img src="crudServices.png" />
    1. `brew install kyleconroy/sqlc/sqlc`, `sqlc help`
    2. [Two configuration versions](https://docs.sqlc.dev/en/stable/reference/config.html#): 1 and 2.
10. **How did SQLC create models for entries and transfers when accounts.sql(inside db/query) only had create statement for Accounts?**
    1. if the accounts.sql file is messed up, an error is thrown, hence not creating the further `.go` files.

# Why DB Migrations?
1. Change of schema - mutliple devs can update the schema at the same time, which schema to keep, which to reject, or if both schemas are right but different due to , say different columns, then how to merge?
2. Change of the engine itself(say from postgres to MySQL)
    1. usual pattern: use SQLite while developing, change to Postgres/MySQL while deploying - implementing this change requires db migration.
3. To Trigger/capture changes in the schema
    1. the 0001_up.sql, 0002_up.sql... and 0001_down.sql, 0002_down.sql... keep the schema dynamic(in code-form , rather than a hardcoded one).
4. Migration is nothing but a *schema creating SQL script*, translating the DB from one state to another.
    1. Say in the [experiment-2 example(migrations-related)](#exp2_db_migration), `first-state = DB{accounts}`, `second-state = DB{accounts, entries}`, `third-state = DB{accounts, entries, transfers}`.
    2. States are denoted by the `schema_migrations` table, with a column called `version`.
    3. Version control for Databases.

# Experiments to be RUN

## Migration-related
1. After the playlist is finished, try making the developing environment a sqlite one, and migrate to a staging env having a postgres DB.
2. <a name="exp2_db_migration"></a>Try to have only `accounts` table at first, and then develop the system further to hold transactions(`entries`) and then eventually to hold `transfers`.

## CRUD Operations
1. once the project is done, have code(in different git branches) corresponding to each of the 4 CRUD services(Database, Gorm, SQLx, SQLc)
