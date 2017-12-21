CREATE TABLE feed_items(
  item_id serial primary key not null,
  feed_id integer not null,
  title varchar(1024) not null,
  content varchar(1024) not null,
  categories varchar(1024) not null,
  description varchar(1024) not null,
  link varchar(1024) not null,
  timestamp timestamp default current_timestamp
);