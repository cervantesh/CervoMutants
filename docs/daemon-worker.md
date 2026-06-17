# Daemon And Worker Mode

`cervomut daemon` and `cervomut worker` currently remain experimental.

They exist so the project can exercise a JSON-lines remote-execution shape
without pretending that the distributed protocol is product-ready.

## Current Status

The mode is available, but it is not a supported compatibility surface yet.

That means:

- the JSON message shape may still change
- there is no protocol version negotiation
- there are no leases or durable work claims
- there are no built-in retries
- there is no supported multi-worker coordination story yet

For that reason, the CLI now requires explicit opt-in:

```powershell
cervomut daemon --experimental
cervomut worker --experimental
```

You can also opt in through:

```powershell
$env:CERVOMUT_EXPERIMENTAL_DAEMON = "1"
```

## What It Does Today

The current implementation is intentionally small:

- reads newline-delimited JSON messages from stdin
- accepts `job` messages containing one mutant job
- executes the job in an isolated copied module
- writes newline-delimited JSON results to stdout

The worker path is useful for:

- internal protocol experiments
- local integration tests
- future distributed-execution spikes

## What It Does Not Promise

It should not yet be presented as:

- a stable remote execution API
- a daemon protocol with backward-compatibility guarantees
- a production queue worker
- a supported CI distribution mode

Those claims require follow-up work such as explicit protocol versioning,
durable coordination, retries, and compatibility policy. That later policy work
is now reflected in [compatibility-policy.md](compatibility-policy.md), which
keeps daemon/worker explicitly outside the supported compatibility surface until
the protocol is versioned and promoted.
