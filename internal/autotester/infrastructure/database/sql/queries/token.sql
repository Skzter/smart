-- name: ReadToken :one
select *
from refresh_tokens
where
user_id = $1;

-- name: CreateToken :one
insert into refresh_tokens(
user_id,
token,
created_at,
updated_at,
expires_at,
revoked_at
)
values (
    $1,
    $2,
    now(),
    now(),
    $3,
    null
)
returning *;

-- name: UpdateToken :one
update refresh_tokens
set
token = $2,
created_at = now(),
updated_at = now(),
expires_at = $3,
revoked_at = $4
where
user_id = $1
returning *;
