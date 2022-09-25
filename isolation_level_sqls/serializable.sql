/* initial table:
 id | owner  | balance | currency |          created_at          
----+--------+---------+----------+------------------------------
 59 | rlkdlb |      89 | CAD      | 2022-09-25 13:14:50.80261+00
 60 | ypwqvd |      12 | EUR      | 2022-09-25 13:14:50.81128+00
 61 | name1  |     200 | USD      | 2022-09-25 13:27:43.059463+00
*/

/* transaction-1 (tx1) */
BEGIN;
set transaction isolation level serializable;
select * from accounts;
/*
 id | owner  | balance | currency |          created_at          
----+--------+---------+----------+------------------------------
 59 | rlkdlb |      89 | CAD      | 2022-09-25 13:14:50.80261+00
 60 | ypwqvd |      12 | EUR      | 2022-09-25 13:14:50.81128+00
 61 | name1  |     200 | USD      | 2022-09-25 13:27:43.059463+00
*/
insert into accounts (owner, balance, currency)values ('sum1', 301, 'USD');
select * from accounts;
/*
 id | owner  | balance | currency |          created_at          
----+--------+---------+----------+------------------------------
 59 | rlkdlb |      89 | CAD      | 2022-09-25 13:14:50.80261+00
 60 | ypwqvd |      12 | EUR      | 2022-09-25 13:14:50.81128+00
 61 | name1  |     200 | USD      | 2022-09-25 13:27:43.059463+00
 62 | sum1   |     301 | USD      | 2022-09-25 13:37:43.059463+00
*/
commit;

/*++++++++++++++++ transaction-2(tx2) START  ++++++++++++++*/
BEGIN;
set transaction isolation level serializable;
select * from accounts; -- ran after the update statement from tx1
/*
 id | owner  | balance | currency |          created_at          
----+--------+---------+----------+------------------------------
 59 | rlkdlb |      89 | CAD      | 2022-09-25 13:14:50.80261+00
 60 | ypwqvd |      12 | EUR      | 2022-09-25 13:14:50.81128+00
 61 | name1  |     200 | USD      | 2022-09-25 13:27:43.059463+00
*/
insert into accounts (owner, balance, currency)values ('sum2', 301, 'USD');
/*
 id | owner  | balance | currency |          created_at          
----+--------+---------+----------+------------------------------
 59 | rlkdlb |      89 | CAD      | 2022-09-25 13:14:50.80261+00
 60 | ypwqvd |      12 | EUR      | 2022-09-25 13:14:50.81128+00
 61 | name1  |     200 | USD      | 2022-09-25 13:27:43.059463+00
 63 | sum2   |     301 | USD      | 2022-09-25 13:37:43.059463+00
*/
/*
notice that psql updates the sequence variable that initializes
a new record's ID field, and communicates it to another session of psql
*/
-- after tx1 is committed:
commit;
/*
ERROR:  could not serialize access due to read/write dependencies among transactions
DETAIL:  Reason code: Canceled on identification as a pivot, during commit attempt.
HINT:  The transaction might succeed if retried.
*/
select * from accounts where balance < 10; /* phantom read AVOIDED*/
/*
 id | owner  | balance | currency |          created_at          
----+--------+---------+----------+------------------------------
 60 | ypwqvd |       7 | EUR      | 2022-09-25 13:14:50.81128+00
*/

/*++++++++++++++++ transaction-2(tx2) END  ++++++++++++++*/

/*
    1. this means that duplicate record creation is also avoided.
*/

