whitelisted relay
=================

  - a basic relay implementation based on relayer.
  - uses postgres, which I think must be over version 12 since it uses generated columns.
  - only accepts events from specific pubkeys defined via the environment variable `WHITELIST` (comma-separated).

running
-------

grab a binary from the releases page and run it with the environment variable POSTGRESQL_DATABASE set to some postgres url:

    POSTGRESQL_DATABASE=postgres://name:pass@localhost:5432/dbname ./relayer-whitelisted

it also accepts a HOST and a PORT environment variables.

compiling
---------

if you know Go you already know this:

    go install github.com/permadao/ArNostr-relayer/whitelisted

or something like that.
