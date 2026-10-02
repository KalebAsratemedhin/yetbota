-- +goose Up
-- Replace activity badges with score-milestone badges (Maps / community style).

DELETE FROM user_badges
WHERE badge_id IN (
    'first_post',
    'ten_likes_received',
    'first_follower',
    'level_contributor',
    'level_neighborhood',
    'level_insider',
    'level_legend'
);

DELETE FROM badges
WHERE id IN (
    'first_post',
    'ten_likes_received',
    'first_follower',
    'level_contributor',
    'level_neighborhood',
    'level_insider',
    'level_legend'
);

INSERT INTO badges (id, name, description) VALUES
    ('level_newcomer', 'Newcomer', 'Getting started on the map'),
    ('level_regular', 'Regular', 'Showing up around town'),
    ('level_contributor', 'Contributor', 'Building to help others discover places'),
    ('level_local_guide', 'Local Guide', 'Trusted voice for what''s around'),
    ('level_explorer', 'Explorer', 'Going further to uncover what''s worth knowing'),
    ('level_legendary_explorer', 'Legendary explorer', 'Widely known for solid local takes'),
    ('level_pathfinder', 'Pathfinder', 'Among the most helpful locals')
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name,
    description = EXCLUDED.description;

-- +goose Down
DELETE FROM user_badges
WHERE badge_id IN (
    'level_newcomer',
    'level_regular',
    'level_contributor',
    'level_local_guide',
    'level_explorer',
    'level_legendary_explorer',
    'level_pathfinder'
);

DELETE FROM badges
WHERE id IN (
    'level_newcomer',
    'level_regular',
    'level_contributor',
    'level_local_guide',
    'level_explorer',
    'level_legendary_explorer',
    'level_pathfinder'
);

INSERT INTO badges (id, name, description) VALUES
    ('first_post', 'First Post', 'Published your first post'),
    ('ten_likes_received', 'Ten Likes', 'Received 10 likes on your posts'),
    ('first_follower', 'First Follower', 'Gained your first follower'),
    ('level_contributor', 'Contributor', 'Reached Contributor reputation level')
ON CONFLICT (id) DO NOTHING;
