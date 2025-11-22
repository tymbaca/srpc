[![Go Reference](https://pkg.go.dev/badge/github.com/tymbaca/srpc.svg)](https://pkg.go.dev/github.com/tymbaca/srpc)

# sRPC - Stupid RPC

Package srpc provides simple and composable RPC implementation, similar to
standard `net/rpc` but with `context.Context` support and type-safe client/server
stub generation tool.

## Getting started

Let's say you have this interface:

```go
package main

type UserService interface {
    Create(ctx context.Context, user User) (Empty, error)
    GetUsers(ctx context.Context, filter UserFilter) ([]User, error)
}

type User struct{
    ID       string
    Name     string
    Password string
}

type UserFilter struct{
    IDs   []string
    Names []string
}

type Empty struct{}
```

Put `go:generate` comment with `srpc-gen --target=UserService` in the same
package, e.g.:

```go
//go:generate srpc-gen --target=UserService
type UserService interface {
    // ...
}
```

Call `go generate` command, specifying the package where your interface is
placed (e.g. `go generate ./...`).

Files `srpc.UserService.client.go` and `srpc.UserService.server.go` will be
generated. They contains type-safe stubs for client and server, both
implementing your target interface.

Now you can use them as you wish. For example:

```go
package main 

import (
    // ...

	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/codec/json"
	"github.com/tymbaca/srpc/transport/stdnet"
)

const addr = "localhost:8080"

func server() {
    server := NewUserServiceClient(srpc.NewServer(json.Codec))
    defer server.Close()

    l, err := stdnet.Listen("tcp", addr)
    if err != nil {
        panic(err)
    }

    _ = server.Start(context.Background(), l)
}

func client() {
    client := NewUserServiceClient(srpc.NewClient(addr, json.Codec, stdnet.NewDialer("tcp")))

    users, err := client.GetUsers(context.Background(), UserFilter{})
    // ...
}
```
