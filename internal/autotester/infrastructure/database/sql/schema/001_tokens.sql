-- +goose Up
create table refresh_tokens(
    id serial primary key,
    user_id text not null, 
    token text not null unique,
    created_at timestamp not null,
    updated_at timestamp not null,
    expires_at timestamp not null,
    revoked_at timestamp
);

create index refresh_tokens_useridtoken_index
on refresh_tokens (user_id, token);

-- +goose Down
drop index refresh_tokens_useridtoken_index;

drop table refresh_tokens;

