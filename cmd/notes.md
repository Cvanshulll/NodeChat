1. listen to all ip in LAN
2. broadcast my LAN IP
3. accept client request

ctx is a tool that helps manage Redis operations, ensuring they can be canceled or timed out when necessary.
Redis acts as a middleman for chat messages, ensuring all users in a chatroom receive messages in real-time.


Client                          Server                          Redis
  |                               |                              |
  |--- Discover Server IP ------->|                              |
  |                               |                              |
  |<-- Broadcast Server IP -------|                              |
  |                               |                              |
  |--- Connect to Server -------->|                              |
  |                               |                              |
  |--- Send Username ------------>|                              |
  |                               |                              |
  |--- Send Chatroom Name -------->|                              |
  |                               |--- Subscribe to Chatroom --->|
  |                               |                              |
  |<-- Welcome Message -----------|                              |
  |                               |                              |
  |--- Send Message --------------|                              |
  |                               |--- Publish Message to Redis >|
  |                               |                              |
  |                               |<-- Redis Sends Message ------|
  |<-- Receive Message -----------|                              |
  |                               |                              |
  |--- Exit Chatroom ------------>|                              |
  |                               |--- Unsubscribe from Chatroom |
  |                               |                              |



+-------------------+          +-------------------+          +-------------------+
|                   |          |                   |          |                   |
|      Client       |          |      Server       |          |       Redis       |
|                   |          |                   |          |                   |
+-------------------+          +-------------------+          +-------------------+
         |                              |                              |
         |--- UDP Broadcast Request --->|                              |
         |                              |                              |
         |<-- UDP Broadcast Response ---|                              |
         |                              |                              |
         |--- TCP Connect -------------->                              |
         |                              |                              |
         |--- Send Username ------------>                              |
         |                              |                              |
         |--- Send Chatroom Name ------->                              |
         |                              |--- Subscribe to Chatroom --->|
         |                              |                              |
         |<-- Welcome Message ----------|                              |
         |                              |                              |
         |--- Send Message ------------>|                              |
         |                              |--- Publish to Redis -------->|
         |                              |                              |
         |                              |  +-----------------------+   |
         |                              |  | Redis Channel: room1  |   |
         |                              |  |-----------------------|   |
         |                              |  | "Alice: Hello!"       |   |
         |                              |  | "Bob: Hi Alice!"      |   |
         |                              |  +-----------------------+   |
         |                              |                              |
         |                              |<-- Redis Sends Message ------|
         |<-- Receive Message ----------|                              |
         |                              |                              |
         |--- Exit Chatroom ------------|                              |
         |                              |--- Unsubscribe from Redis -->|
         |                              |                              |
         |--- Disconnect ---------------|                              |
         |                              |                              |