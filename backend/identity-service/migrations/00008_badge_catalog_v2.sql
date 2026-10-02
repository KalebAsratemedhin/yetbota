-- +goose Up
-- Align catalog with final 7 score badges (drop neighborhood/insider/legend; add explorer + legendary explorer).

DELETE FROM user_badges
WHERE badge_id IN (
    'level_neighborhood',
    'level_insider',
    'level_legend'
);

DELETE FROM badges
WHERE id IN (
    'level_neighborhood',
    'level_insider',
    'level_legend'
);

INSERT INTO badges (id, name, description) VALUES
    ('level_explorer', 'Explorer', 'Going further to uncover what''s worth knowing'),
    ('level_legendary_explorer', 'Legendary explorer', 'Widely known for solid local takes')
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name,
    description = EXCLUDED.description;

-- +goose Down
DELETE FROM user_badges
WHERE badge_id IN (
    'level_explorer',
    'level_legendary_explorer'
);

DELETE FROM badges
WHERE id IN (
    'level_explorer',
    'level_legendary_explorer'
);

INSERT INTO badges (id, name, description) VALUES
    ('level_neighborhood', 'Neighborhood voice', 'People notice your takes on nearby spots'),
    ('level_insider', 'Local insider', 'Deeply useful on what''s worth knowing'),
    ('level_legend', 'Neighborhood legend', 'Widely known for solid local takes')
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name,
    description = EXCLUDED.description;
