-- CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, user_id, content)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;
