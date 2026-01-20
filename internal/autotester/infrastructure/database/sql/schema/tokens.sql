create table refresh_tokens(
    id serial primary key,
    user_id text not null unique,
    token text not null unique,
    created_at timestamp not null,
    updated_at timestamp not null,
    expires_at timestamp not null,
    revoked_at timestamp
);
