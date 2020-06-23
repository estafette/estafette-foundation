# Estafette

The `estafette-foundation` library is part of the Estafette CI system documented at https://estafette.io.

Please file any issues related to Estafette CI at https://github.com/estafette/estafette-ci-central/issues

## Estafette-foundation

This library provides building blocks for creating

This library has contracts for requests / responses between various components of the Estafette CI system.

## Development

To start development run

```bash
git clone git@github.com:estafette/estafette-foundation.git
cd estafette-foundation
```

Before committing your changes run

```bash
go test ./...
go mod tidy
```

## Usage

To add this module to your golang application run

```bash
go get github.com/estafette/estafette-foundation
```

### Initialize logging

```go
import "github.com/estafette/estafette-foundation"

foundation.InitLogging(app, version, branch, revision, buildDate)
```

### Initialize Prometheus metrics endpoint

```go
import "github.com/estafette/estafette-foundation"

foundation.InitMetrics()
```

### Handle graceful shutdown

```go
import "github.com/estafette/estafette-foundation"

gracefulShutdown, waitGroup := foundation.InitGracefulShutdownHandling()

// your core application logic, making use of the waitGroup for critical sections

foundation.HandleGracefulShutdown(gracefulShutdown, waitGroup)
```


### Watch mounted folder for changes

```go
import "github.com/estafette/estafette-foundation"

foundation.WatchForFileChanges("/path/to/mounted/secret/or/configmap", func(event fsnotify.Event) {
  // reinitialize parts making use of the mounted data
})
```

### Apply jitter to a number to introduce randomness

Inspired by http://highscalability.com/blog/2012/4/17/youtube-strategy-adding-jitter-isnt-a-bug.html you want to add jitter to a lot of parts of your platform, like cache durations, polling intervals, etc.

```go
import "github.com/estafette/estafette-foundation"

sleepTime := foundation.ApplyJitter(30)
time.Sleep(time.Duration(sleepTime) * time.Second)
```

### Retry

In order to retry a function you can use the `Retry` function to which you can pass a retryable function with signature `func() error`:

```go
import "github.com/estafette/estafette-foundation"

foundation.Retry(func() error { do something that can fail })
```

Without passing any additional options it will by default try 3 times, with exponential backoff with jitter applied to the interval for any error returned by the retryable function.

In order to override the defaults you can pass them in with the following options:

```go
import "github.com/estafette/estafette-foundation"

foundation.Retry(func() error { do something that can fail }, Attempts(5), DelayMillisecond(10), Fixed())
```

The following options can be passed in:

| Option   | Config property | Description |
| -------- | --------------- | ----------- |
| Attempts | Attempts        | Sets the number of attempts the retryable function will be attempted before returning the error |
| DelayMillisecond | DelayMillisecond | Sets the base number of milliseconds between the retries or to base the exponential backoff delay on |
| ExponentialJitterBackoff | DelayType |
| ExponentialBackoff | DelayType |
| Fixed | DelayType |
| AnyError | IsRetryableError |

#### Custom options

You can also override any of the config properties by passing in a custom option with signature `func(*RetryConfig)`, which could look like:

```go
import "github.com/estafette/estafette-foundation"

isRetryableErrorCustomOption := func(c *RetryConfig) {
  c.IsRetryableError = func(err error) bool {
    switch e := err.(type) {
      case *googleapi.Error:
        return e.Code == 429 || (e.Code >= 500 && e.Code < 600)
      default:
        return false
    }
  }
}

foundation.Retry(func() error { do something that can fail }, isRetryableErrorCustomOption)
```