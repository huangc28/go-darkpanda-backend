BEGIN;
    ALTER TABLE chatrooms 
    DROP COLUMN IF EXISTS "enabled";
COMMIT;