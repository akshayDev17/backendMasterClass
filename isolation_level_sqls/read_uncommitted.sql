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
 59 | rlkdlb |     105 | CAD      | 2022-09-25 13:14:50.80261+00
 60 | ypwqvd |       7 | EUR      | 2022-09-25 13:14:50.81128+00
*/
insert into accounts(owner, balance, currency) values ('name1', 200, 'USD');
select * from accounts;
/*
id | owner  | balance | currency |          created_at          
----+--------+---------+----------+------------------------------
 59 | rlkdlb |     105 | CAD      | 2022-09-25 13:14:50.80261+00
 60 | ypwqvd |       7 | EUR      | 2022-09-25 13:14:50.81128+00
 61 | name1  |     200 | USD      | 2022-09-25 13:27:43.059463+00
*/

/* transaction-2(tx2)  */
BEGIN;
set transaction isolation level read uncommitted;
select * from accounts; -- ran after the update statement from tx1
/*
id | owner  | balance | currency |          created_at          
----+--------+---------+----------+------------------------------
 59 | rlkdlb |     105 | CAD      | 2022-09-25 13:14:50.80261+00
 60 | ypwqvd |       7 | EUR      | 2022-09-25 13:14:50.81128+00
*/
/*
this shows that in psql , READ UNCOMMITTED functions like how a 
READ COMMITTED isolation level should function as.
*/
