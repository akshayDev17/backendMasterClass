/* initial table:
id | owner  | balance | currency |          created_at          
----+--------+---------+----------+------------------------------
 59 | rlkdlb |     105 | CAD      | 2022-09-25 13:14:50.80261+00
 60 | ypwqvd |       7 | EUR      | 2022-09-25 13:14:50.81128+00
 61 | name1  |     200 | USD      | 2022-09-25 13:27:43.059463+00
*/

/* transaction-1 (tx1) */
BEGIN;
set transaction isolation level repeatable read;
select * from accounts;
/*
id | owner  | balance | currency |          created_at          
----+--------+---------+----------+------------------------------
 59 | rlkdlb |      89 | CAD      | 2022-09-25 13:14:50.80261+00
 60 | ypwqvd |       7 | EUR      | 2022-09-25 13:14:50.81128+00
 61 | name1  |     200 | USD      | 2022-09-25 13:27:43.059463+00
*/
update accounts set balance = balance + 5 where id = 60;
select * from accounts;
/*
id | owner  | balance | currency |          created_at          
----+--------+---------+----------+------------------------------
 59 | rlkdlb |      89 | CAD      | 2022-09-25 13:14:50.80261+00
 60 | ypwqvd |      12 | EUR      | 2022-09-25 13:14:50.81128+00
 61 | name1  |     200 | USD      | 2022-09-25 13:27:43.059463+00
*/
commit;

/*++++++++++++++++ transaction-2(tx2) START  ++++++++++++++*/
BEGIN;
set transaction isolation level repeatable read;
select * from accounts; -- ran after the update statement from tx1
/*
 id | owner  | balance | currency |          created_at          
----+--------+---------+----------+------------------------------
 59 | rlkdlb |      89 | CAD      | 2022-09-25 13:14:50.80261+00
 60 | ypwqvd |       7 | EUR      | 2022-09-25 13:14:50.81128+00
 61 | name1  |     200 | USD      | 2022-09-25 13:27:43.059463+00
*/
select * from accounts where balance < 10;
/*
 id | owner  | balance | currency |          created_at          
----+--------+---------+----------+------------------------------
 60 | ypwqvd |       7 | EUR      | 2022-09-25 13:14:50.81128+00
*/

-- after tx1 is committed:
select * from accounts where balance < 10; /* phantom read AVOIDED*/
/*
 id | owner  | balance | currency |          created_at          
----+--------+---------+----------+------------------------------
 60 | ypwqvd |       7 | EUR      | 2022-09-25 13:14:50.81128+00
*/
select * from accounts; /* REPEATABLE READ */
/*
 id | owner  | balance | currency |          created_at          
----+--------+---------+----------+------------------------------
 59 | rlkdlb |      89 | CAD      | 2022-09-25 13:14:50.80261+00
 60 | ypwqvd |       7 | EUR      | 2022-09-25 13:14:50.81128+00
 61 | name1  |     200 | USD      | 2022-09-25 13:27:43.059463+00
*/

update accounts set balance = balance - 5 where id = 60 returning *;
/*
ERROR:  could not serialize access due to concurrent update
*/

/*++++++++++++++++ transaction-2(tx2) END  ++++++++++++++*/

/*

*/

