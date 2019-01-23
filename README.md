Couchbase integration testing in Go
===================================

`gocbt` is a package to help you write integration tests with Couchbase. If your Go project uses Couchbase and you want to write integration tests for it, `gocbt` can simplify the setup and teardown of nodes.

By default `gocbt` pulls the [official Couchbase Docker image](https://hub.docker.com/r/couchbase/server/), but can be configured to pull any image from any repository. Configuration of the nodes is completely customizable, but sensible defaults are used so only the settings important to your particular use case need be defined.

Below is a simple demonstration of how `gocbt` is used with default configuration. Read about what configuration options are available [here](https://godoc.org/github.com/joe-mann/gocbt).

```go
func TestUserRegistration(t *testing.T) {
    node := gocbt.NewNode()
    defer node.Teardown(t)
    node.Setup(t)
    node.Configure(t, gocbt.Bucket("users"))

    // Connect to Couchbase
    app := StartApp(Config{
        CouchbaseHost:     node.Host(),
        CouchbaseUsername: node.Username(),
        CouchbasePassword: node.Password(),
    })

    // Run the test case
    err := RegisterUser("John", "Doe", "john.doe@domain.com")
    if err != nil {
        t.Fatal("failed to register user")
    }
}
```

`gocbt` currently only targets Couchbase Server 6.0.0 Community; other versions may work but have not been tested.