# Inquiry Domain

## Male emits inquiry

Male user can emit a new inquiry via `POST inquiries`. This API would perform following things:

    - Check if the male user already has an active inquiry.
    - If active inquiry exists and not expired, returns error response.
    - If the inquiry is expired mark that inquiry to be expired. 
    - Create a new inqiury.
    - Male user joins lobby.

### Join lobby

當男性用戶加入大廳，我們會新增一筆紀錄到 DB `lobby_users` table 中。然後我們會使用這筆紀錄的 `uuid` 來新增一筆新的 document 到 firestore `lobby collection` 中。這個動作的意義在於當女生 pick up inquiry 時，我們可以透過修改這筆 `lobby document` 來通知男性用戶有女生 pick 了他的 inquiry。而如果這個男性用戶點選 `馬上聊聊` ，此筆 `lobby document` 的狀態會被修改，女生被通知狀態後會進入到 chatroom 開始進行聊天。

`firestore` 在這裡的角色就像 `socket`，用來通知男女雙方 lobby 的狀態，然後進行相對應的處理:

  - 女方 pickup inquiry， 男生跳 popup
  - 男方按下 `馬上聊聊` 女方 inquiry tab 中多一筆 inquiry，點擊此筆 inquiry 則進入到聊天室
  - 男方按下略過，此筆 lobby document 的狀態回到 `waiting`

### Pickup Inquiry

`/inquiries/:inquiry_uuid/pickup` 

女性用戶可以 pick up 一個 inquiry。pick up 後，會改變 DB 中此 `lobby_user` 的狀態為 `asking`。表示有女性正在詢問中。也會把 firestore 此 lobby user 的狀態改為 asking。這時男生的手機就會被通知了。

## TODOs

- [x] Get inquiries. Query condition should exclude expired records. We should add `expired_at` column in `lobby_users` table and remove `expired_at` in `lobby_users` table.
- [x] Cancel inquiry.
- [x] User revert inquiry from chatting. Soft both delete user's presence in that chatroom and the chatroom itself. If is a male user, rejoining lobby if inquiry isn't expired.
- [] Reverting inquiry should emit `quit` message to members in the channel.
- [] When service provider picks up an inquiry, backend should create a private chat document in firestore. Client will then subscribe to this document and the subcollections beneath this document.

## Bugs

- [] If a service provider is already chatting with the customer, she should not be able to pickup the same inquiry again.
- [x] If an inquiry is at `chatting` status, it should be removed to lobby. Other service providers should not be able to find the inquiry.
- [x] Error #01: sql: transaction has already been committed or rolled back.

Committed / Roll backed transaction should be replaced with normal SQL client. A better solution would be to resolve a whole new DAO Service object on each request to avoid reusing Roll backed / Committed transactions.
