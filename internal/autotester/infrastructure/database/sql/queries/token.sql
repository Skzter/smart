-- name: GetTokenByUserID :one
select token
from refresh_tokens
where
user_id = $1 and 
expires_at > now() and
revoked_at is null and 
token is not null;

-- name: InsertTokenByUserID :exec
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
    now() + interval '24 hours',
    null
);

-- name: UpdateTokenByUserID :exec
update refresh_tokens
set
token = $2,
updated_at = now()
where
user_id = $1 and
expires_at > now() and
revoked_at is null;
