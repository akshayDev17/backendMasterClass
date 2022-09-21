# Table of Contents
1. [Developmental Steps](#developmental_steps)
    1. [Unit Test](#unit_test)
2. [SQLC Build](#sqlc_build)
3. [How does `sqlc init` work?](#sqlc_init)
4. [How does `sqlc generate` work?](#sqlc_gen)
5. [Why DB Migrations?](#why_db_migration)
6. [Experiments to be RUN](#experiments)

# Developmental Steps<a name="developmental_steps"></a>
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
    Ways to implement CRUD:\ 
    <img src="crudServices.png" />
    1. DB/SQL [QueryRowContext](https://pkg.go.dev/database/sql#example-DB.QueryRowContext) 
    2. Gorm's functions refers to the CRUD Interface functions mentioned [here](https://gorm.io/docs/create.html). You would need to also use the functions in the Associations section to make gorm understand the associations between various fields.
    3. sql/x - extension on db/sql so that struc/ slice can be formed from results of SQL queries using functions such as [Get, Select, StructScan](https://pkg.go.dev/github.com/jmoiron/sqlx#readme-usage).
    4. SQL/C - read the first 3 steps of the [documentation](https://docs.sqlc.dev/en/stable/).
        1. Since SQL/C processes these queries supplied to it, any errors will now be thrown at **compile time**.
        2. **Note**: MySQL is now supported(image is old).
        3. Even this is an extension on `database/sql` (observe `accounts.sql.go`).
    2. `brew install kyleconroy/sqlc/sqlc`, `sqlc help`
10. Instead of using the `go mod init` approach, we created folders inside `db(sqlc)` which will house the go files created by `sqlc`.
    1. `sqlc init` was run before directory creation, to create the `sqlc.yaml` config file,using either of the [Two configuration versions](https://docs.sqlc.dev/en/stable/reference/config.html#): 1 and 2.
    2. `make sqlc` will create the db/{accounts.sql.go,db.go,models.go} files - anytime this command is executed, the **files** that **exist earlier** are **overwritten**.
    3. `db/query/accounts.sql` ---> `db/sqlc/accounts.sql.go`, which now contains custom golang functions for creating an account, listing accounts, fetching a particular account, because these were the `SQLs` defined in `accounts.sql`.
    4. ```go
        return &Queries{db: db}
        ``` 
        this is a pointer notation of creating a struct(`Queries` is the struct here) in Go, with db field = db passed as an argument to the function `func New(db DBTX) *Queries`. \
        Note(could be removed after polishing this README): we can have structs that implement only certain functions of an interface.
    5. A variable declared as a particular type of interface can trigger any of the methods of this interface that are implemented by this specific data-type.
        ```go
        package main
        import (
            "fmt"
        )

        type geometry interface {
            area() float64
            perim() float64
        }
        type rect struct {
            width, height float64
        }

        func (r rect) area() float64 { // interface methods
            return r.width * r.height // implemented for rect struct
        }

        type circle struct {
            radius float64
        }

        func (c circle) area() float64 { // interface methods
            return math.Pi * c.radius * c.radius // implemented for circle struct
        }
        func (c circle) perim() float64 {
            return 2 * math.Pi * c.radius
        }

        func measure(g geometry) {
            fmt.Println(g)
            fmt.Println(g.area())
            fmt.Println(g.perim())
        }

        func main() { 
            r := rect{width: 3, height: 4}
            c := circle{radius: 5}
            fmt.Println(r.area())
        }
        ```
        1. `measure(r)` will fail because there is no implementation of `perim()` method for the `struct rect`, whereas `measure(c)` will be sucessful.
11. [`Meanings of :one, :many, :exec annotations`](https://docs.sqlc.dev/en/stable/reference/query-annotations.html) w.r.t. the `accounts.sql` file.
    1. Not necessary to add `LIMIT` and `OFFSET` for the `:many` SQL Annotation.
    2. 
12. **How did SQLC create models for entries and transfers when accounts.sql(inside db/query) only had create statement for Accounts?**
    1. if the `accounts.sql` file is messed up, an error is thrown, hence not creating the further `.go` files.
13. ## Unit Test for CRUD Operations<a name="unit_test"></a>
    1. Golang convention - put the test file(`account_test.go`) in the same folder as the code.
    2. created main_test.go to house the connection object to DB
        1. [`lib/pq`]() was installed since we lack a psql driver to establish a connection, the `database/sql` only provides a Golang-based interface to communicate with a pre-existing driver. \
            Also mentioned in [Getting Started with Postgres in sqlc](https://docs.sqlc.dev/en/stable/tutorials/getting-started-postgresql.html).
        2. `go get lib/pq` will  fail as installing packages using go get is deprecated, use go install instead.
        3. Hence, a go mod file was created:
            ```bash
            go mod init my/backendMasterClass
            go get github.com/lib/pq # adds github.com/lib/pq as required package in go.mod file
            go install # installs all packages specified in go.mod
            ```
        4. Running Test Functions:
            ```bash
            go test -run '' # all tests
            go test -run ^TestMain # run all functions in all .go files in the current directory having name like TestMain%
            ```
            1. if the import `_ "github.com/lib/pq"` is commented, then the test will fail, since it doesn't have the driver to connect to.
            2. `TestMain` is itself [not a testing function](https://medium.com/goingogo/why-use-testmain-for-testing-in-go-dafb52b406bc), it rather runs all the Testing functions in the current directory(when `m.Run()` is called).
        5. **GOLANG NAMING CONVENTION**: functions having `TestXXYY` as the name(starting with Test) and taking `testing.T` as an argument will be referred to as unit tests by golang, and will be always run unless otherwise.
    3. *Testing C(create) operation* out of CRUD:
        1. Created `account_test.go` to create a dummy account each time the `TestAccount` function is run.
            1. created `util.go` belonging to package `util` which would use random generators for each field of `Account`.
            2. Observe what is the module name in go.mod (here its my/backendMasterclass). \
            Hence when util is to be imported into account_test.go, the path would be `<module_name>/<package_name>`
        2. For checking test results, [stretchr/testify](https://github.com/stretchr/testify/tree/v1.8.0) package was used.
            1. The require object was used to check for integrity and type constraints of each field of the created Account object.
        3. Runnin test from .: `go test my/backendMasterclass/db/sqlc -cover -v -run TestCreateAccount^` (because test files are in db/sqlc)
    4. *Testing R(retrieve/fetch/GET) operation* out of CRUD:
        1. In `account_test.go`, added the function `TestGetAccount`
        2. created a new account and fetched its detail into another account variable
        3. compared them attribute wise using `require`.
    5. *Testing U(update) operation*
        1. Updated a newly created account.
        2. Notice standard definition of update is `:exec` , we have given `:one`, and additionally `RETURNING *` , in `db/query/accounts.sql`
    6. *Testing D(delete) operation*
        1. `DeleteAccount` has again `:exec`
        2. `require.Error` is used for the first time, as we want to assert that the deleted account's ID shouldn't exist in the table.
14. DB Transactions
    1. Need for them
        1. Unit of work should be reliable and consistent, even during system failures.
        2. isolate programs that access DB at the same time.
    2. Example - *Transfer 500USD from Rob's Account(ID=123) to Bill's Account(ID=456)*
        1. Fetch account1 and account2(verifying existence)
        2. Check whether account1 and account2 have the same currencies.
        3. Check whether account1's balance is \>= 500
        4. Create a transfer record with from id = account1.ID, to_account = account2.ID , amount=500
        5. Create 2 entry records
            1. accountID = account1.ID, amount=-500
            2. accountID = account2.ID, amount=+500
        6. Update balance of both accounts.
    3. **ACID**
        1. *Atomicity*: all operations(unit tasks) of a transaction complete successfully or none do.
        2. *Consistency*: DB state must be valid(all constraints hold).
        3. *Isolation*: concurrent transactions shouldn't affect wach other.
        4. *Durability*: A successful transaction should write data into persistent storage.
    4. the `execTx` function is unexported, i.e. used only within the same package, [reference](https://stackoverflow.com/questions/40256161/exported-and-unexported-fields-in-go-language).
    5. [sqlc reference](https://docs.sqlc.dev/en/stable/howto/transactions.html) for using transactions.

# SQLC Build<a name="sqlc_build"></a>
1. `Dockerfile` runs `go run scripts/release.go -docker`.
2. The `if *docker` snippet then runs the go build command via `exec.Command`(same as `os.system` in Python)
    1. The entire command = `go build -a -ldflags -extldflags \"-static\" -X github.com/kyleconroy/sqlc/internal/cmd.version=1.8.0 -o /workspace/sqlc ./cmd/sqlc` (here I've put the latest version manually, but the snippet before this determines the actual version).
3. `go build` creates a binary, which can then be [executed as](https://gobyexample.com/command-line-arguments) : `go build fileName.go ; ./fileName`.
    1. [meanings](https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies) of all valid flags with `go build`.
    2. force rebuild all packages(`-a`), pass the flags listed in `-ldflags` to all future runs of [`go link`](https://pkg.go.dev/cmd/link) (this will refer to a file [src/cmd/cgo/doc.go](https://go.dev/src/cmd/cgo/doc.go), whose line 1024 is relevant) 
        1. the flags passed to -ldflags are 
            1. `-extldflags`
                1. The space-separated strings following this are flags passed to the external linker.
                2. using this means that an external linker is used.
                3. Here, its `-static`
                    1. static libraries [utilized](https://www.youtube.com/watch?v=-vp9cFQCQCo) at compile time
                        1. as opposed to shared ones, which are utilized at run time, hence for them we need to have the app and the library at the same time.
                        2. whereas for static libraries, the entire code in them is copied into the app at compile time, hence only the app is needed.
                    2. hence the `-extldflags -static` means to use an external linker(`ld` in this case) and to only link static libraries of go.
            2. `-X github.com/kyleconroy/sqlc/internal/cmd.version=1.8.0` - set the importpath's name to this github go package.
            3. [linking dependencies in static libraries](https://stackoverflow.com/questions/7841920/how-do-static-libraries-do-linking-to-dependencies)
    3. `-o workspace/sqlc` : the output binary of this `go build` run should be stored in this path.
        1. Notice in the dockerfile, . is copied into /workspace and we then switch to [WORKDIR](https://www.educative.io/answers/what-is-the-workdir-command-in-docker), which is /workspace.(equivalent of `mkdir`+`cd`)
    4. the [complete command for go build](https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies) is `go build [-o output] [build flags] [packages]`
        1. the `./cmd/sqlc` is the package being asked to compile + run.
        2. use `go help packages`:
        > Usually, \[packages\] is a list of import paths. \
        Import paths beginning with "cmd/" only match source code in the Go repository.
4. We now `cd` into `workspace/sqlc`.

# How does `sqlc init` work?<a name="sqlc_init"></a>
1. Observe the [`cmd/sqlc`](https://github.com/kyleconroy/sqlc/blob/v1.8.0/internal/cmd/cmd.go) package.
    1. Observe the `initCmd` variable, which uses another package called [`cobra`](https://github.com/spf13/cobra/tree/v1.5.0) containing [commands](https://github.com/spf13/cobra/blob/v1.5.0/command.go).
        1. Compare the values of Use, Short and Run with the outputs of `sqlc init -h`.
        2. The Flag and Flags and FlagSet are all borrowed from a repo forked by [`spf13`](https://github.com/spf13/pflag/tree/v1.0.5) from ogier/pflag.
        3. [`cmd.Flags()`](https://github.com/spf13/cobra/blob/v1.5.0/command.go#L1466) --> [`c.flags`](https://github.com/spf13/cobra/blob/06b06a9dc9f9f5eba93c552b2532a3da64ef9877/command.go#L133) is a FlagSet pointer(pointing to a list of flags)
            1. this will be nil for the very first run but after that will be initialized.
            2. [`FlagSet`](https://github.com/spf13/pflag/blob/v1.0.5/flag.go#L138) from ogier/pflag
            3. [`Lookup`](https://github.com/spf13/pflag/blob/v1.0.5/flag.go#L348) method implemented by the FlagSet struct.
            4. 
    2. 

# How does `sqlc generate` work?<a name="sqlc_gen"></a>

# Why DB Migrations?<a name="why_db_migration"></a>
1. Change of schema - mutliple devs can update the schema at the same time, which schema to keep, which to reject, or if both schemas are right but different due to , say different columns, then how to merge?
2. Change of the engine itself(say from postgres to MySQL)
    1. usual pattern: use SQLite while developing, change to Postgres/MySQL while deploying - implementing this change requires db migration.
3. To Trigger/capture changes in the schema
    1. the `0001_up.sql`, `0002_up.sql`... and `0001_down.sql`, `0002_down.sql`... keep the schema dynamic(in code-form , rather than a hardcoded one).
4. Migration is nothing but a *schema creating SQL script*, translating the DB from one state to another.
    1. Say in the [experiment-2 example(migrations-related)](#exp2_db_migration), `first-state = DB{accounts}`, `second-state = DB{accounts, entries}`, `third-state = DB{accounts, entries, transfers}`.
    2. States are denoted by the `schema_migrations` table, with a column called `version`.
    3. Version control for Databases.

# Experiments to be RUN<a name="experiments"></a>

## Migration-related
1. After the playlist is finished, try making the developing environment a sqlite one, and migrate to a staging env having a postgres DB.
2. <a name="exp2_db_migration"></a>Try to have only `accounts` table at first, and then develop the system further to hold transactions(`entries`) and then eventually to hold `transfers`.

## CRUD Operations
1. once the project is done, have code(in different git branches) corresponding to each of the 4 CRUD services(Database, Gorm, SQLx, SQLc)
