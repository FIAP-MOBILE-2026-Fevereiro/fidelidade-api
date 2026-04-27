INSERT INTO programs (
    id,
    merchant_id,
    merchant_name,
    lat,
    lng,
    stamp_goal,
    reward_name,
    reward_image_url,
    reward_description,
    description,
    rules,
    active,
    starts_at,
    ends_at
) VALUES (
    'prog_b4c5d6e7',
    'est_a1b2c3d4',
    'Café do Centro',
    -23.550520,
    -46.633308,
    10,
    'Café Expresso Grátis',
    'https://storage.fidelidadeapp.com/recompensas/cafe_gratis.jpg',
    'Um delicioso café expresso 100% arábica',
    'A cada café, ganhe um selo!',
    'Válido para compras acima de R$15,00. Máximo 1 selo por dia.',
    TRUE,
    now() - interval '30 days',
    now() + interval '180 days'
) ON CONFLICT (id) DO NOTHING;

INSERT INTO qr_codes (
    id,
    program_id,
    merchant_id,
    code_hash,
    raw_payload,
    expires_at
) VALUES (
    'qr_a1b2c3d4',
    'prog_b4c5d6e7',
    'est_a1b2c3d4',
    encode(digest('eyJwcm9ncmFtYUlkIjoicHJvZ19iNGM1ZDZlNyIsImVzdGFiZWxlY2ltZW50b0lkIjoiZXN0X2ExYjJjM2Q0IiwiY2FyaW1ibyI6IjEyMzQ1Njc4OTAiLCJ0aW1lc3RhbXAiOiIyMDI2LTA0LTI3VDE0OjMwOjAwWiJ9', 'sha256'), 'hex'),
    'eyJwcm9ncmFtYUlkIjoicHJvZ19iNGM1ZDZlNyIsImVzdGFiZWxlY2ltZW50b0lkIjoiZXN0X2ExYjJjM2Q0IiwiY2FyaW1ibyI6IjEyMzQ1Njc4OTAiLCJ0aW1lc3RhbXAiOiIyMDI2LTA0LTI3VDE0OjMwOjAwWiJ9',
    now() + interval '7 days'
) ON CONFLICT (id) DO NOTHING;