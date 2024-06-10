create table user_state (
    user_id     integer primary key,
    state       text not null,
    foreign key(user_id) references user(user_id)
);