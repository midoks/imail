# go-rspamd

## Introduction <br/>

go-rspamd is a client library written to help interact with a rspamd instance, via HTTP. go-rspamd facilitates
* Content scanning of emails
* Training rspamd's Bayesian classifier
* Updating rspamd's fuzzy storage by adding or removing messages 
* Analyzing content scanning results 

Refer to rspamd [documentation](https://rspamd.com/doc/) for help configuring and setting up rspamd.

## Usage 

The API is defined [here](https://pkg.go.dev/github.com/Shopify/go-rspamd).

The client helps send emails to all POST endpoints on rspamd. Support for all GET endpoints does not currently exist as they can be accessed through rspamd's web interface, although support may exist in the future. A full list of all endpoints can be found [here](https://rspamd.com/doc/architecture/protocol.html). 

The client supports email formats that implement `io.Reader` or `io.WriterTo`. For example, this means that clients can pass in both `gomail.Message` objects, which implement `io.WriteTo` or simply the contents of an `.eml` file, which implement `io.Reader`. gomail can be found [here](https://github.com/go-gomail/gomail).

### Examples

_Note:_ go-rspamd is geared towards clients that use [context](https://golang.org/pkg/context/). However if you don't, whenever `context.Context` is expected, you can use `context.Background()`.

Import go-rspamd:
```go
import "github.com/Shopify/go-rspamd"
```

Instantiate the client with the url of your rspamd instance:
```go
 client := rspamd.New("https://contentscanner.com")
```

Optionally pass in credentials:
```go
 client := rspamd.New("https://contentscanner.com", rspamd.Credentials("username", "password"))
```

Ping your rspamd instance:
```go
pong, _ := client.Ping(ctx)
```

Scan an email from an io.Reader (eg. loading an `.eml` file):
```go
f, _ := os.Open("/path/to/email")
email := rspamd.NewEmailFromReader(f).QueueId(2)
checkRes, _ := client.Check(ctx, email)
```

Scan an email from an io.WriteTo (eg. a gomail `Message`):
```go
// let mail be of type *gomail.Message
// attach a Queue-Id to rspamd.Email instance
email := rspamd.NewEmailFromWriterTo(mail).QueueID(1)
checkRes, _ := client.Check(ctx, email)
```

Add a message to fuzzy storage, attaching a flag and weight as per [docs](https://rspamd.com/doc/architecture/protocol.html#controller-http-endpoints):
```go
// let mail be of type *gomail.Message
email := rspamd.NewEmailFromWriterTo(mail).QueueID(2).Flag(1).Weight(19)
learnRes, _ := client.FuzzyAdd(ctx, email)
```

## Semantics

### Contributing

We invite contributions to extend the API coverage.

Report a bug: Open an issue  
Fix a bug: Open a pull request

### Versioning

go-rspamd respects the [Semantic Versioning](https://semver.org/) for major/minor/patch versions. 

### License

MIT
