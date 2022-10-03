create table if not exists orderdb.orders (
    id int not null auto_increment,
    order_id varchar(32) not null,
    item_id varchar(32) not null,
    item_name varchar(32) not null,
    price integer not null,
    token_address varchar(32) not null,
    token_id smallint not null,
    delivered boolean not null,
    primary key (id)
)