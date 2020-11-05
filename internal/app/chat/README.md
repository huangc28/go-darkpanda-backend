## TODOs

- [x] Emit message to private chatroom.
- [x] Initialize pubnub instance. Pubnub instance should be singleton.
- [x] Add validators before emitting message to private chat.
- [x] Retrieve the latest message from each chat from firestore. This message is used to display in the inquiry chatroom tab page. 
- [] The message retireved from the above requirement should be order by created_at timestamp. 
- [] Move chat_dao from `inquiry` domain to `chat` domain.

