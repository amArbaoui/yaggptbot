create table "user" (
    user_id     integer primary key,
    tg_user_id  integer not null,
    tg_chat_id  integer not null,
    tg_username text not null,
    created_at  integer not null,
    updated_at     integer null
);

create table "role" (
    role text primary key not null
);

create table message (
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

create table user_promt (
    user_id     integer not null,
    promt       text null,
    created_at  integer not null,
    updated_at  integer null,
    foreign key(user_id) references user(user_id)
)
