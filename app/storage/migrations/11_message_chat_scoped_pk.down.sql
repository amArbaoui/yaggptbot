create table message_old (
    chat_id         integer not null,
    message_id      integer primary key,
    reply_id        integer null,
    message_text    text not null,
    role            text not null,
    created_at      integer not null,
    updated_at      integer null,
    foreign key(chat_id) references user(tg_chat_id),
    foreign key(role) references role(role)
);

insert or ignore into message_old (chat_id, message_id, reply_id, message_text, role, created_at, updated_at)
select chat_id, message_id, reply_id, message_text, role, created_at, updated_at from message order by created_at;

drop table message;

alter table message_old rename to message;

create index if not exists idx_message_msg_id on message (message_id);
