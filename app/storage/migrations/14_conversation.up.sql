create table conversation (
    chat_id                 integer primary key,
    reset_after_message_id  integer not null
);
