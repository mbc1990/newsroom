CREATE TABLE feeds(
  feed_id serial primary key not null,
  feed_name varchar(1024) not null,
  feed_link varchar(1024) not null
);
