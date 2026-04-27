-- name: ListStampsByProgram :many
SELECT
    st.id,
    st.user_id,
    st.program_id,
    st.acquired_at,
    st.validation_key,
    st.sequence,
    st.validated,
    p.merchant_name
FROM stamps st
JOIN programs p ON p.id = st.program_id
WHERE st.user_id = $1
  AND st.program_id = $2
ORDER BY st.sequence ASC;

-- name: CountStampsByProgram :one
SELECT COUNT(*)::int
FROM stamps
WHERE user_id = $1
  AND program_id = $2;

-- name: GetMaxStampSequence :one
SELECT COALESCE(MAX(sequence), 0)::int
FROM stamps
WHERE user_id = $1
  AND program_id = $2;

-- name: FindStampByQRCodeHash :one
SELECT id
FROM stamps
WHERE qr_code_hash = $1;

-- name: CreateStamp :one
INSERT INTO stamps (
    id,
    user_id,
    program_id,
    qr_code_id,
    qr_code_hash,
    validation_key,
    sequence
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
RETURNING id, user_id, program_id, acquired_at, validation_key, sequence, validated;

-- name: GetQRCodeValidation :one
SELECT
    q.id,
    q.program_id,
    q.merchant_id,
    q.code_hash,
    q.raw_payload,
    q.generated_at,
    q.expires_at,
    q.used,
    q.used_by_user_id,
    q.used_at,
    p.merchant_name,
    p.stamp_goal,
    p.reward_name,
    p.reward_image_url,
    p.reward_description
FROM qr_codes q
JOIN programs p ON p.id = q.program_id
WHERE q.code_hash = $1;

-- name: MarkQRCodeUsed :one
UPDATE qr_codes
SET used = TRUE,
    used_by_user_id = $2,
    used_at = now()
WHERE id = $1
  AND used = FALSE
RETURNING id;