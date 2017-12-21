CREATE TABLE feed_items(
  item_id serial primary key not null,
  feed_title varchar(1024) not null,
  title varchar(1024) not null unique,
  content text not null,
  description text not null,
  link text not null,
  timestamp timestamp default current_timestamp,
  scraped boolean not null
);
