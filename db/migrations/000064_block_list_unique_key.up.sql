CREATE UNIQUE INDEX idx_userid_blocked_userid ON block_list(user_id, blocked_user_id);