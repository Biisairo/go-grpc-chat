# go-grpc-chat

### build buf stub
```buf generate```

### http header to meta data
http header
```Grpc-Metadata-user_id: 123```

grpc metadata
```"user_id": "123"```

### http gateway feature
- unary - possible
- serer streaming - possible
- client streaming - impossible
- bidi streaming - impossible
