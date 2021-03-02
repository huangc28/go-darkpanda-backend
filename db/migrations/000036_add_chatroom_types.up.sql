CREATE TYPE chatroom_type AS ENUM(
	'inquiry_chat',
	'service_chat'
);

BEGIN;

ALTER TABLE chatrooms
ADD COLUMN chatroom_type chatroom_type NOT NULL DEFAULT 'inquiry_chat';

COMMIT;
