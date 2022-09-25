/* initial table:
id | owner  | balance | currency |          created_at          
----+--------+---------+----------+------------------------------
 59 | rlkdlb |     105 | CAD      | 2022-09-25 13:14:50.80261+00
 60 | ypwqvd |       7 | EUR      | 2022-09-25 13:14:50.81128+00
*/

/* transaction-1 (tx1) */
BEGIN;
set transaction isolation level read uncommitted;
select * from accounts;
/*
id | owner  | balance | currency |          created_at          
----+--------+---------+----------+------------------------------
 59 | rlkdlb |     100 | CAD      | 2022-09-25 13:14:50.80261+00
 60 | ypwqvd |       7 | EUR      | 2022-09-25 13:14:50.81128+00
*/
update accounts set balance = balance - 11 where id = 59;
select * from accounts;
/*
id | owner  | balance | currency |          created_at          
----+--------+---------+----------+------------------------------
 59 | rlkdlb |      89 | CAD      | 2022-09-25 13:14:50.80261+00
 60 | ypwqvd |       7 | EUR      | 2022-09-25 13:14:50.81128+00
 61 | name1  |     200 | USD      | 2022-09-25 13:27:43.059463+00
*/
commit;

/*++++++++++++++++ transaction-2(tx2) START  ++++++++++++++*/
BEGIN;
set transaction isolation level read uncommitted;
select * from accounts; -- ran after the update statement from tx1
/*
id | owner  | balance | currency |          created_at          
----+--------+---------+----------+------------------------------
 59 | rlkdlb |     100 | CAD      | 2022-09-25 13:14:50.80261+00
 60 | ypwqvd |       7 | EUR      | 2022-09-25 13:14:50.81128+00
*/
select * from accounts where balance > 90;
/*
 id | owner  | balance | currency |          created_at          
----+--------+---------+----------+------------------------------
 59 | rlkdlb |     100 | CAD      | 2022-09-25 13:14:50.80261+00
*/

-- after tx1 is committed:
select * from accounts where balance > 90; /* 0 rows - PHANTOM READ*/
select * from accounts; /* NON-REPEATABLE READ */
/*
id | owner  | balance | currency |          created_at          
----+--------+---------+----------+------------------------------
 59 | rlkdlb |     89  | CAD      | 2022-09-25 13:14:50.80261+00
 60 | ypwqvd |       7 | EUR      | 2022-09-25 13:14:50.81128+00
*/
/*++++++++++++++++ transaction-2(tx2) END  ++++++++++++++*/

/*
select * from accounts; in tx2 experiences a NON-REPEATABLE READ as the
row with id=59 now has different value of balance.(due to tx1 commit)

select * from accounts where balance > 90; in tx2 experiences a PHANTOM READ as the
as earlier 1 row was returned, now 0 rows are returned.(due to tx1 commit)
*/

