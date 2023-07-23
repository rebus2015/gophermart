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

alter table users
    owner to pguser;

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
    accural  bigint    default 0 not null,
    date_ins timestamp default now(),
    constraint orders_pk
        primary key (user_id, num)
);

alter table orders
    owner to pguser;

create table if not exists  withdraws
(
    user_id  uuid                    not null
        constraint withdraws_fk
            references users
            on delete cascade,
    num      bigint                  not null,
    expence  bigint                  not null,
    date_ins timestamp default now() not null,
    constraint withdraws_pk
        primary key (user_id, num)
);

alter table withdraws
    owner to pguser;

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
           ELSE 0::numeric
           END AS accs,
       CASE
           WHEN (EXISTS (SELECT
                         FROM withdraws w
                         WHERE w.user_id = u.id)) THEN (SELECT sum(w.expence) AS sum
                                                        FROM withdraws w
                                                        WHERE w.user_id = u.id)
           ELSE 0::numeric
           END AS exps
FROM users u;

alter table user_balance
    owner to pguser;
create function user_add(_login character varying, _hash bytea) returns character varying
    language sql
as
$$
   INSERT INTO users (id, hash, login)
values (gen_random_uuid (),_hash,_login)
ON CONFLICT on constraint users_un
do nothing
returning cast(id as varchar);
$$;

alter function user_add(varchar, bytea) owner to pguser;

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

alter function public.user_check(varchar, out varchar, out bytea) owner to pguser;

create function orders_all(_user_id uuid)
    returns TABLE(num bigint, status character varying, accural bigint, date_ins timestamp without time zone)
    language sql
as
$$
SELECT  num, status, accural, date_ins
FROM orders
where user_id  = _user_id
order by date_ins asc;
$$;

alter function orders_all(uuid) owner to pguser;

create function withdraw_add(_user_id uuid, _number bigint, _expence bigint) returns bigint
    language sql
as
$$
INSERT INTO withdraws (user_id, num, expence, date_ins)
values (_user_id, _number, _expence, default)
on conflict on constraint withdraws_pk
do nothing
returning num;
$$;

alter function withdraw_add(uuid, bigint, bigint) owner to pguser;

create function withdrawals_all(_user_id uuid)
    returns TABLE(num bigint, expence bigint, date_ins timestamp without time zone)
    language sql
as
$$
 select num, expence, date_ins from withdraws
 where user_id = _user_id
 order by date_ins asc
$$;

alter function withdrawals_all(uuid) owner to pguser;

create function order_add(_user_id uuid, _number bigint, _status character varying, _accural bigint) returns SETOF uuid
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
return query (select user_id from orders where num = _number);
else
 begin
	insert
	into orders (user_id, num, status,	accural, date_ins)
    values (_user_id, _number, _status, _accural, default);
 end;
end if;
end;

$$;

alter function order_add(uuid, bigint, varchar, bigint) owner to pguser;

create function balance(_user_id uuid)
    returns TABLE(balance bigint, expence bigint)
    language sql
as
$$
SELECT  (b.accs - b.exps) as balance, exps as expence
FROM user_balance b
where b.id  = _user_id
$$;

alter function balance(uuid) owner to pguser;

-- +goose StatementEnd