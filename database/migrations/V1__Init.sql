create table if not exists orderdb.orders (
    order_id varchar(64) not null,
    item_id varchar(64) not null,
    item_name varchar(64) not null,
    price bigint not null,
    delivery_price bigint not null,
    token_address varchar(64) not null,
    token_id smallint not null,
    delivered boolean not null,
    primary key (order_id)
)