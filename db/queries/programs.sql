-- name: CountNearbyPrograms :one
SELECT COUNT(*)::bigint
FROM programs p
WHERE p.active = TRUE
  AND p.starts_at <= now()
  AND (p.ends_at IS NULL OR p.ends_at >= now())
  AND ST_DWithin(
    ST_SetSRID(ST_MakePoint(p.lng, p.lat), 4326)::geography,
    ST_SetSRID(ST_MakePoint(sqlc.arg(lng), sqlc.arg(lat)), 4326)::geography,
    sqlc.arg(radius_meters)
    );

-- name: ListNearbyPrograms :many
SELECT
    p.id,
    p.merchant_name,
    p.lat,
    p.lng,
    ST_Distance(
        ST_SetSRID(ST_MakePoint(p.lng, p.lat), 4326)::geography,
        ST_SetSRID(ST_MakePoint(sqlc.arg(lng), sqlc.arg(lat)), 4326)::geography
    )::double precision AS distance,
    p.stamp_goal,
    COALESCE((
        SELECT COUNT(*)::int
        FROM stamps st
        WHERE st.program_id = p.id
                    AND st.user_id = sqlc.arg(user_id)
    ), 0)::int AS current_stamps,
    p.reward_name,
    p.reward_image_url,
    p.reward_description,
    p.description,
    p.rules,
    p.active,
    p.starts_at,
    p.ends_at
FROM programs p
WHERE p.active = TRUE
  AND p.starts_at <= now()
  AND (p.ends_at IS NULL OR p.ends_at >= now())
  AND ST_DWithin(
        ST_SetSRID(ST_MakePoint(p.lng, p.lat), 4326)::geography,
                ST_SetSRID(ST_MakePoint(sqlc.arg(lng), sqlc.arg(lat)), 4326)::geography,
                sqlc.arg(radius_meters)
    )
ORDER BY distance ASC, p.created_at DESC
LIMIT sqlc.arg(limit_rows) OFFSET sqlc.arg(offset_rows);

-- name: GetProgramDetail :one
SELECT
    p.id,
    p.merchant_name,
    p.lat,
    p.lng,
    p.stamp_goal,
    COALESCE((
        SELECT COUNT(*)::int
        FROM stamps st
        WHERE st.program_id = p.id
          AND st.user_id = $2
    ), 0)::int AS current_stamps,
    p.reward_name,
    p.reward_image_url,
    p.reward_description,
    p.description,
    p.rules,
    p.active,
    p.starts_at,
    p.ends_at
FROM programs p
WHERE p.id = $1;

-- name: ListActiveUserPrograms :many
SELECT
    p.id AS program_id,
    p.merchant_name,
    p.stamp_goal,
    COUNT(st.id)::int AS obtained_stamps,
    GREATEST(p.stamp_goal - COUNT(st.id), 0)::int AS remaining_stamps,
    LEAST(100, FLOOR((COUNT(st.id)::numeric / p.stamp_goal) * 100))::int AS progress,
    MAX(st.acquired_at)::timestamptz AS last_stamp_at,
    (COUNT(st.id) = p.stamp_goal - 1) AS almost_complete,
    p.reward_name,
    p.reward_image_url,
    p.reward_description
FROM programs p
JOIN stamps st ON st.program_id = p.id
WHERE st.user_id = $1
  AND p.active = TRUE
  AND (p.ends_at IS NULL OR p.ends_at >= now())
GROUP BY p.id
HAVING COUNT(st.id) > 0 AND COUNT(st.id) < p.stamp_goal
ORDER BY MAX(st.acquired_at) DESC;

-- name: ListFinishedUserPrograms :many
WITH progress AS (
    SELECT
        p.id AS program_id,
        p.merchant_name,
        p.stamp_goal,
        COUNT(st.id)::int AS obtained_stamps,
        MAX(st.acquired_at) AS last_stamp_at,
        p.ends_at,
        p.reward_name,
        p.reward_image_url,
        p.reward_description,
        rr.redeemed,
        rr.redeemed_at,
        rr.completed_at
    FROM programs p
    JOIN stamps st ON st.program_id = p.id
    LEFT JOIN reward_redemptions rr ON rr.program_id = p.id AND rr.user_id = $1
    WHERE st.user_id = $1
    GROUP BY p.id, rr.redeemed, rr.redeemed_at, rr.completed_at
), classified AS (
    SELECT
        program_id,
        merchant_name,
        reward_name,
        reward_image_url,
        reward_description,
        CASE
            WHEN redeemed THEN 'resgatado'
            WHEN obtained_stamps >= stamp_goal THEN 'completo'
            WHEN ends_at IS NOT NULL AND ends_at < now() THEN 'expirado'
            ELSE NULL
        END::text AS status,
        COALESCE(completed_at, ends_at, last_stamp_at) AS finished_at,
        redeemed,
        redeemed_at
    FROM progress
    WHERE obtained_stamps >= stamp_goal
       OR (ends_at IS NOT NULL AND ends_at < now())
)
SELECT
    program_id,
    merchant_name,
    reward_name,
    reward_image_url,
    reward_description,
    status,
    finished_at,
    redeemed,
    redeemed_at
FROM classified
WHERE status IS NOT NULL
  AND (sqlc.arg(filter_status)::text = '' OR status = sqlc.arg(filter_status)::text)
ORDER BY finished_at DESC;

-- name: GetLastProgramStamp :one
SELECT
    p.id AS program_id,
    p.merchant_name,
    p.stamp_goal,
    totals.obtained_stamps,
    st.id AS stamp_id,
    st.user_id,
    st.program_id AS stamp_program_id,
    st.acquired_at,
    st.validation_key,
    st.sequence,
    st.validated
FROM stamps st
JOIN programs p ON p.id = st.program_id
JOIN LATERAL (
    SELECT COUNT(*)::int AS obtained_stamps
    FROM stamps sx
    WHERE sx.user_id = st.user_id
      AND sx.program_id = st.program_id
) totals ON TRUE
WHERE st.user_id = $1
ORDER BY st.acquired_at DESC
LIMIT 1;

-- name: ListAvailableRewards :many
SELECT
    rr.program_id,
    p.merchant_name,
    p.reward_name,
    p.reward_image_url,
    p.reward_description,
    rr.completed_at,
    rr.expires_at
FROM reward_redemptions rr
JOIN programs p ON p.id = rr.program_id
WHERE rr.user_id = $1
  AND rr.redeemed = FALSE
  AND (rr.expires_at IS NULL OR rr.expires_at > now())
ORDER BY rr.completed_at DESC;

-- name: GetRewardRedemptionByProgram :one
SELECT id, user_id, program_id, reward_code, completed_at, expires_at, redeemed, redeemed_at, created_at
FROM reward_redemptions
WHERE user_id = $1
  AND program_id = $2;

-- name: UpsertRewardRedemption :one
INSERT INTO reward_redemptions (
    id,
    user_id,
    program_id,
    reward_code,
    completed_at,
    expires_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
ON CONFLICT (user_id, program_id) DO UPDATE
SET user_id = EXCLUDED.user_id
RETURNING id, user_id, program_id, reward_code, completed_at, expires_at, redeemed, redeemed_at, created_at;

-- name: RedeemReward :one
UPDATE reward_redemptions
SET redeemed = TRUE,
    redeemed_at = now()
WHERE user_id = $1
  AND program_id = $2
  AND reward_code = $3
  AND redeemed = FALSE
RETURNING redeemed_at;