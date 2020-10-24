## TODOs

- [x] Get inquiries. Query condition should exclude expired records. We should add `expired_at` column in `lobby_users` table and remove `expired_at` in `lobby_users` table.
- [] Cancel inquiry 
- [] User revert inquiry from chatting. Soft delete user's presence in that chatroom and the chatroom itself. If is a male user, rejoining lobby if inquiry isn't expired.
- [] Reverting inquiry should emit `quit` message to members in the channel.