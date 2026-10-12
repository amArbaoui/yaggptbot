create table message_new (
    chat_id         integer not null,
    message_id      integer not null,
    reply_id        integer null,
    message_text    text not null,
    role            text not null,
    created_at      integer not null,
    updated_at      integer null,
    primary key (chat_id, message_id),
    foreign key(chat_id) references user(tg_chat_id),
    foreign key(role) references role(role)
);

insert into message_new (chat_id, message_id, reply_id, message_text, role, created_at, updated_at)
select chat_id, message_id, reply_id, message_text, role, created_at, updated_at from message;

drop table message;

alter table message_new rename to message;
