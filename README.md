# drop-con

This is an application designed to mishandle HTTP requests.

There are four endpoints exposed:

- `/hi` - responds with a `200 OK` and the message "Hi!"
- `/drop` - accepts the request and then silently drops the connection
- `/dropNoisy` - accepts the request, starts responding and then drops the connection
- `/hang` - holds the connection open indefinitely

## Running on Cloud Foundry

Build the binary.

```
GOOS=linux GOARCH=amd64 go build -o app
```

Push to the platform.

```
cf push
```