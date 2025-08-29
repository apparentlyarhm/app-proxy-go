# APP-PROXY-GO

A go-based backend service that connects to `Github`, `Steam` and `Spotify` to return data to my frontend, [arhm.dev](https://arhm.dev) (The application isnt live yet)

### info

so instead of it being a 1-to-1 rewrite, I have tried my best to learn best practices and implement it here.

- **Services as clients**: What we've done here is abstract each service as a client. Each "client" has a cfg associated with it, which holds env vars and stuff. We define various methods (that are essentially API calls to these services) as methods of the struct `Client`. hence all methods can share the cfg. So,
**package -> client -> method[1..n]**. This pattern is very similar to java.

- **The "Server" Struct Pattern**: We avoid global state by encapsulating all our application's dependencies (clients) into a single Server struct. Our HTTP handlers are methods on this struct, giving them clean, controlled access to these dependencies.

### THIS IS MOSTLY A RE-WRITE OF MY APP IN EXPRESS, FOR EDUCATIONAL PURPOSES.

## log

- [29/8/25]: Finished porting steam API integration