DROP TABLE IF EXISTS entries;
DROP TABLE IF EXISTS transfers;
DROP TABLE IF EXISTS accounts;
/*
the order of dropping tables matter, because entries and transfers have account_id related
foreign key constraint-dependency on accounts table.
hence while migrating down if accounts is dropped first,
entries and transfers though existent will not have those constraints
satisfied, hence will throw an error.
*/