-- name: ReadToken :one
select *
from refresh_tokens
where
user_id = $1;

-- name: UpsertToken :one
insert into refresh_tokens (
    user_id,
    token,
    created_at,
    updated_at,
    expires_at,
    revoked_at
) values (
    $1,
    $2,
    now(),
    now(),
    $3,
    $4
)
on conflict (user_id) 
do update set
    token = excluded.token,
    created_at = now(),
    updated_at = now(),
    expires_at = excluded.expires_at,
    revoked_at = excluded.revoked_at
returning *;
