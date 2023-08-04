-- +goose Up
-- +goose StatementBegin

create table if not exists  users
(
    id    uuid    not null
        constraint users_pk
            primary key,
    hash  bytea   not null,
    login varchar not null
        constraint users_un
            unique
);

create unique index users_id_idx
    on users (id);

create table if not exists  orders
(
    user_id  uuid                not null
        constraint orders_fk
            references users
            on delete cascade,
    order_id bigint generated always as identity
        constraint orders_un
            unique,
    num      bigint              not null,
    status   varchar             not null,
    accural  numeric(9,2)    default 0 not null,
    date_ins timestamp default now(),
    constraint orders_pk
        primary key (user_id, num)
);

create table if not exists  withdraws
(
    user_id  uuid                    not null
        constraint withdraws_fk
            references users
            on delete cascade,
    num      bigint                  not null,
    expence  numeric(9,2)                  not null,
    date_ins timestamp default now() not null,
    constraint withdraws_pk
        primary key (user_id, num)
);


CREATE OR REPLACE VIEW user_balance(id, accs, exps) as
SELECT u.id,
       CASE
           WHEN (EXISTS (SELECT
                         FROM orders o
                         WHERE o.user_id = u.id
                           AND o.status::text = 'PROCESSED'::text)) THEN (SELECT sum(o.accural) AS sum
                                                                          FROM orders o
                                                                          WHERE o.status::text = 'PROCESSED'::text
                                                                            AND o.user_id = u.id)
           ELSE 0::numeric(9,2)
           END AS accs,
       CASE
           WHEN (EXISTS (SELECT
                         FROM withdraws w
                         WHERE w.user_id = u.id)) THEN (SELECT sum(w.expence) AS sum
                                                        FROM withdraws w
                                                        WHERE w.user_id = u.id)
           ELSE 0::numeric(9,2)
           END AS exps
FROM users u;


create or replace function user_add(_login character varying, _hash bytea) returns character varying
    language sql
as
$$
   INSERT INTO users (id, hash, login)
values (gen_random_uuid (),_hash,_login)
ON CONFLICT on constraint users_un
do nothing
returning cast(id as varchar);
$$;

create or replace function public.user_check(_login character varying, OUT id character varying, OUT hash bytea) returns record
    language sql
as
$$
   select
       cast(u.id as varchar) as id,
       u.hash AS hash
   from users u
   where login =_login
    $$;

create or replace function orders_all(_user_id uuid)
    returns TABLE(num bigint, status character varying, accural numeric(9,2), date_ins timestamp without time zone)
    language sql
as
$$
SELECT  num, status,
        case when status='PROCESSED' THEN accural ELSE 0::numeric(9,2) END,
        date_ins
FROM orders
where user_id  = _user_id
order by date_ins asc;
$$;

create or replace function withdraw(_user_id uuid, _number bigint, _expence numeric(9,2)) returns boolean
    language plpgsql
as
$$
declare 
    cur numeric(9,2);
begin
cur := (select b.balance from balance(_user_id) as b);
if (cur > _expence)
then
    begin       
       INSERT INTO withdraws (user_id, num, expence, date_ins)
        values (_user_id, _number, _expence, default)
        on conflict on constraint withdraws_pk
        do nothing;
       return true;
    end;
    else return false;
end if;
end;
$$;

create or replace function withdrawals_all(_user_id uuid)
    returns TABLE(num bigint, expence numeric(9,2), date_ins timestamp without time zone)
    language sql
as
$$
 select num, expence, date_ins from withdraws
 where user_id = _user_id
 order by date_ins asc
$$;

create or replace function order_add(_user_id uuid, _number bigint, _status character varying, _accural numeric(9,2)) returns SETOF text
    language plpgsql
as
$$
begin
if exists(
select
from
	orders
where
	num = _number)
then
return query (select cast(user_id as text) from orders where num = _number);
else
 begin
	insert
	into orders (user_id, num, status,	accural, date_ins)
    values (_user_id, _number, _status, _accural, default);
    return query (select '' as user_id);
 end;
end if;
end;
$$;

create or replace function balance(_user_id uuid)
    returns TABLE(balance numeric(9,2), expence numeric(9,2))
    language sql
as
$$
SELECT  (b.accs - b.exps) as balance, exps as expence
FROM user_balance b
where b.id  = _user_id
$$;



create or replace function order_update(_num bigint, _status character varying, _accrual numeric(9,2)) returns void
    language sql
as
$$
update orders set
                  status = _status,
                  accural = _accrual
    WHERE num = _num
$$;

create function orders_acc()
    returns TABLE(num bigint, status character varying)
    language sql
as
$$
SELECT  num, status
FROM orders
where status not in ('PROCESSED','INVALID')
$$;

-- +goose StatementEnd