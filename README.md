# jsonbank-cli

Command-line interface for [JSONBank](https://jsonbank.io) — store, fetch, and manage JSON documents.

## Install

**Homebrew** (macOS / Linux)

```sh
brew install jsonbankio/tap/jsb
```

**Go**

```sh
go install github.com/jsonbankio/jsonbank-cli/cmd/jsb@latest
```

**Binaries**

Prebuilt binaries for macOS, Linux, and Windows are on the [releases page](https://github.com/jsonbankio/jsonbank-cli/releases).

## Usage

Log in with your JSONBank API keys, then work with files:

```sh
jsb auth login
```

```sh
jsb file view <path>              # print a file's JSON
jsb file download <path> <dest>   # save a file locally
jsb file meta <path>              # show a file's metadata
jsb file update <idOrPath> <file> # update a document from a local JSON file
```

### Accounts

Multiple accounts can be saved and switched between:

```sh
jsb auth whoami            # show the active account
jsb auth accounts list     # list saved accounts
jsb auth accounts add      # add an account without making it active
jsb auth switch            # switch the active account
jsb auth accounts remove <username>
```

Run `jsb --help` or `jsb <command> --help` for details on any command.

## Development

```sh
make build   # build ./jsb
make test    # run tests
make install # symlink jsb into your PATH
```
