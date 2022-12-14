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
            Also mentioned in [Getting Started with Postgres in sqlc](https://docs.sqlc.dev/en/stable/tutorials/getting-started-postgresql.html). \
            [ALSO MENTIONED IN THE SQL DATABASE DRIVERS LIST FOR GOLANG](https://github.com/golang/go/wiki/SQLDrivers)
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
    5. While updating the balances, a new sql is defined in the `db/sqlc/accounts.sql` file called `AddAccountBalance`, checkout how we add `sqlc.arg`.
    5. [sqlc reference](https://docs.sqlc.dev/en/stable/howto/transactions.html) for using transactions.
15. SQL Update and Select happen concurrently, without waiting for other updates from other routines.
    1. Hence launching multiple go routines as concurrent transactions might cause race conditions leading to improper data.
    2. <img src="goroutines.png" />
    3. Before comitting, no updates are actually made to the database, hence routine with id=36 reads balances before routine-id=34 updates the changes post transactions.
    4. This doesn't mean each incoming transaction needs to be sequential, only those that **access the same data**.
    5. Hence such transactions should acquire a lock which will prevent others from even running their read operation(`SELECT *`).
    6. Rows matching the WHERE condition are locked by an incoming routine.
        1. Hence From account ID will be locked by id=34, but id=35 blocks the To Account ID rows, thus causing a deadlock.
    7. [Locking Clause for PostgreSQL](https://www.postgresql.org/docs/current/sql-select.html)(of which UPDATE is a part of).
        1. Using the queries to check which process(concurrent transaction) acquires what locks, its actually concluded that proc-1 acquires locks for `transfers` and proc-2 for `accounts`.
            1. [Psql lock monitoring](https://wiki.postgresql.org/wiki/Lock_Monitoring) \
            Enlist processes that get blocked.
            ```sql
            SELECT 
            blocked_locks.pid     AS blocked_pid,
            blocked_activity.usename  AS blocked_user,
            blocking_locks.pid     AS blocking_pid,
            blocking_activity.usename AS blocking_user,
            blocked_activity.query    AS blocked_statement,
            blocking_activity.query   AS current_statement_in_blocking_process
            FROM  pg_catalog.pg_locks         blocked_locks
            JOIN pg_catalog.pg_stat_activity blocked_activity  ON blocked_activity.pid = blocked_locks.pid
            JOIN pg_catalog.pg_locks         blocking_locks 
            ON blocking_locks.locktype = blocked_locks.locktype
            AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
            AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
            AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
            AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
            AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
            AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
            AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
            AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
            AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
            AND blocking_locks.pid != blocked_locks.pid
            JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
            WHERE NOT blocked_locks.granted;
            ```
            2. The sql query to cause the deadlock will work only in an interactive session with the DB , since we need to BEGIN a transaction.(not possible in TablePlus)
            3. The following query lists all the lock types currently distributed amongst which processes.
            ```sql
            SELECT 
            l.relation::regclass,
            l.transactionid,
            l.mode,
            l.locktype,
            l.GRANTED,
            --          a.usename,
            a.query,
            --          a.query_start,
            --          age(now(), a.query_start) AS "age",
            a.pid
            FROM pg_stat_activity a
            JOIN pg_locks l ON l.pid = a.pid
            where a.application_name='psql'
            ORDER BY a.pid;
            ```
            3. The blocked query is trying to access a `transactionid` type lock but a lock with this transactionid value is already with `pid=1274` as an `ExclusiveLock`.
        2. Even though these tables are different, they are related by the foreign key constraints, hence require the postgres to acquire a `SharedLock`.
        3. Solution: the shared locks will not be acquired if Postgres is ensured that `ID` in `accounts` table will not be changed(when a row of `accounts` is being fetched), thus implying the fact that records in `transfers` will not be interfered with while `accounts` is updated.
        4. `NO KEY UPDATE` is the proper Locking clause to be used here, w.r.t. `GetAccount()`.
    8. Even after the above solution, we have deadlock between two transactions between same accounts if the first involves money transferred from `account1` and the second involves money transferred to `account1`.
        1. This is a classic case of Hold-n-Wait.
        ```sql
        -- first transaction
        BEGIN;
        UPDATE accounts SET balance = balance + 10 WHERE id = 1 RETURNING *;
        UPDATE accounts SET balance = balance - 10 WHERE id = 2 RETURNING *; -- this will be blocked by the first statement of 2nd transaction

        -- second transaction
        BEGIN;
        UPDATE accounts SET balance = balance + 10 WHERE id = 2 RETURNING *;
        UPDATE accounts SET balance = balance - 10 WHERE id = 1 RETURNING *;
        ```
        2. Solution: Keep the update order the same across all transactions. We know accountID is an int64 , hence always update the smaller ID first, then the larger ID.
16. Transaction Isolation Level - [MySQL]
    1. ## Read Phenomena
        1. Dirty Read - data written by uncommitted transaction is read by another transaction.
        2. Non-repeatable read - before and after committing of some other transaction, this transaction reads the same rows with different value(s), due to updates made by that other transaction.
        3. Phantom read - rows satisfying a condition in a transaction are no longer returned after another transaction is committed, due to updates made by that other transaction.
        4. Serialization Anomaly - a result of a group of concurrent transactions couldn't be achieved if the individual transactions are run in any sequence without overlapping each other.
17. Transaction Isolation Level - [PSQL](https://www.postgresql.org/docs/current/transaction-iso.html)
    1. `set transaction isolation level repeatable read;`
        1. If tx1 inserts a new row, until it commits, tx2 doesn't see it.
        2. Now if tx2 inserts a new row before tx1(with inserted row but uncommitted), and then both are committed one after the other, we see both records entered into accounts.
            1. What if the *same record was inserted by tx1 and tx2*(everything except id is the same) --> **duplication of records** ?
        3. This is called ***Serialization Anomaly*** - meaning if we had run and committed tx2 before tx1, the result would've been the same.
    2. **Note:** Transaction isolation level should be set first before calling any other query in that transaction.
    3. Hence implement retry mechanisms.
    4. <img src="psqlIsolations.png" />
18. Github Action for Continuous Integrations
    1. Each integration of a new piece of code to the original code is verified via automated build and test.
    2. What is Workflow, Job, Step, Github Actions?
        1. *workflow*: made of jobs, can be triggered by an event/scheduled trigger/manual trigger. each `.yml` file in `.github/workflows` is a workflow. \
        `on` keyword defines the trigger, in our case `ci.yml` has `push to branches master` event and also a schedule trigger.
        2. *jobs* - each runner for each job, `jobs:` specifies the list of jobs to be run, runner-> github/self hosted \
        `runs-on`is the runner for each specified job. github hosted `ubuntu-latest` used for running.\
        jobs can be run parallely if not dependent on each other, else run sequentially. `needs` is used to specify this dependency.
        3. *steps* - each job is comprised of individual tasks run sequentially.
        4. *actions* - standalone command, run sequentially within a step such as `install golang migrate` contains 3 actions. **Actions can be reused**.
    3. Configure workflow for Go.(github repo-->actions tab)
        1. creates a new file .github/workflows/go.yml
    4. [Github Action docs](https://docs.github.com/en/actions)
        1. `ci.yml` tells us that even `go.sum` is required for managing dependencies.
            1. Difference between `run` and `uses`?(while defining a step)
        2. this tells us that the github action was unable to connect to postgres DB.
            <img src="setupPostgresConnectionCI.png" />
        3. [Github Action for Postgres](https://docs.github.com/en/actions/using-containerized-services/creating-postgresql-service-containers)
            1. The env variables are defined w.r.t. postgres docker.
        4. Migrate needed to be installed. indentation very important w.r.t. the commands given in `run` action.(| helps you specify multiple commands to be run as part of that step)

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

# Context in Input and output HTTP requests<a name="context"></a>
1. `context.Background()` --> returns `background` --> from `new(emptyCtx)` which is a type int having many supplementary functions implemented on it.
2. `Context` is rather an interface having `Deadline, Done, Err, Value` as the methods to be implemented.

# sql methods

## `sql.Open()`<a name="sql_Open"></a>
- needs driverName and dataSourceName.
    - in our testDB case, we give this as `postgres` and the complete uri to our `simple_bank` db. 
- the last line of the docstring of this function reads
    > // The returned DB is safe for concurrent use by multiple goroutines \
    // and maintains its own pool of idle connections. Thus, the Open \
    // function should be called just once. It is rarely necessary to \
    // close a DB. 
- `driversMu.RLock()` is triggered
    - `driversMu` is a variable of type `sync.RWMutex`(read-write mutex, like the reader-writer's problem)
        - it is initialized such that all fields are 0.
    - initially Enabled(an enum from the race package) is set to false, hence code goes to `atomic.AddInt32`
    - the hashmap drivers is also empty initialized, but *somewhere it gets filled with the key-value pair `postgres-<postgres/driver/address>`*
    - Register() needs to be called w.r.t. a driver so that this driver object can be fed into the map.
        - for instance, the `sql_test.go` and `fakedb_test.go` calls the `Register()` function in their `init()` function bodies.
        - *probably  the lib/pq package could call the Register function*
            - > When they say 'import side effects' they are essentially referring to code/features that are used statically. \
            Meaning just the import of the package will cause some code to execute on app start putting my system in a state \
            different than it would be without having imported that package (like code in an init() which in their example \
            registers handlers, it could also lay down config files, modify resource on disc, ect).
            - [`init()`](https://go.dev/doc/effective_go#init) function is triggered as explained in Effective Go.
            - this `init()` function in [lib/pq/conn.go](https://github.com/lib/pq/blob/d5affd5073b06f745459768de35356df2e5fd91d/conn.go#L59) registers the postgres driver.
            - `lib/pq` implements its own Driver class on the basis of the `Driver interface` provided in `database/sql/driver/driver.go` , and the `Open()` interface method is implemented on [this line of `conn.go`](https://github.com/lib/pq/blob/d5affd5073b06f745459768de35356df2e5fd91d/conn.go#L55).
            - for some reason, map assignment gives us an address, but in itself it returns an empty struct.
            ```go
            type MyDriver interface {
                MyOpen(name string) (int64, error)
            }

            type MyDriverStruct struct{}

            var (
                mydrivers = make(map[string]MyDriver)
            )

            func (d MyDriverStruct) MyOpen(name string) (int64, error) {
                return int64(10), nil
            }

            func myRegister(name string, driver MyDriver) {
                mydrivers[name] = driver
            }

            func main() {
                fmt.Println("Before anything:", mydrivers) // Before anything: map[]
                myRegister("myDBName", &MyDriverStruct{})
                fmt.Println("After assignment:", mydrivers) // After assignment: map[myDBName:0x10139f430]
                fmt.Println(reflect.TypeOf(mydrivers["myDBName"])) // *main.MyDriverStruct

                var x = &MyDriverStruct{}
	            fmt.Println(reflect.TypeOf(x), "\t", *x, "\t", x) // *main.MyDriverStruct     {}      &{}
            }
            ```
- the hashmap drivers is checked
    - key = `string`, value = object belonging to the `driver.Driver` interface.

## `db.conn()`
1. obtain mutex lock on `db` of type struct `DB`.
    1. read about atomic functions [here](https://pkg.go.dev/sync/atomic).
    2. mutex 2 properties: state and sema(semaphore)
    3. For testing purposes, in main_test.go, we had defined a `testQueries` object of type `Queries` struct.
        1. its corresponding testDB was defined using [`testDB, err = sql.Open(dbDriver, dbSource)`](#sql_Open)
### `cachedOrNewConn` as the strategy
1. 
### `alwaysNewConn` as the strategy
1. 

## `QueryRow()`
1. defined in `go/1.19/libexec/src/database/sql/sql.go`
2. implemented for all types: `DB, Tx, Conn`.
3. [`background`](#context) context is used.
3. returns `*Rows`
4. executes the `.query()` method.
    1. `cachedOrNewConn` connReuseStrategy was used, which is a const.
    2. is executed twice(`const maxBadConnRetries = 2`), failing this, query is used with `alwaysNewConn` `connReuseStrategy`.


## `QueryRowContext()`
1. returns sql.Row type(struct with `error` and `*Rows` as properties)
2. defined in `go/1.19/libexec/src/database/sql/sql.go`
3. implemented for all types: `DB, Tx, Conn`.

# Experiments to be RUN<a name="experiments"></a>

## Migration-related
1. After the playlist is finished, try making the developing environment a sqlite one, and migrate to a staging env having a postgres DB.
2. <a name="exp2_db_migration"></a>Try to have only `accounts` table at first, and then develop the system further to hold transactions(`entries`) and then eventually to hold `transfers`.

## CRUD Operations
1. once the project is done, have code(in different git branches) corresponding to each of the 4 CRUD services(Database, Gorm, SQLx, SQLc)
