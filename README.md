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

To install directly on a server (no Go required), download, extract, and move onto your `PATH` in one line — swap in the latest version and your architecture (`amd64` or `arm64`):

```sh
curl -sL https://github.com/jsonbankio/jsonbank-cli/releases/download/v0.1.1/jsb_0.1.1_linux_amd64.tar.gz | tar -xz jsb && sudo mv jsb /usr/local/bin/
```

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

## CI

`jsb` reads API keys from the environment when they are set:

```sh
JSB_PUBLIC_KEY=...
JSB_PRIVATE_KEY=...
```

These take priority over any saved account, so CI needs no `jsb auth login` step — set the variables and run commands directly. If either variable is set, both keys come from the environment; saved keys are ignored entirely.

Example — a GitHub Actions job that pushes a JSON file to JSONBank:

```yaml
jobs:
  publish:
    runs-on: ubuntu-latest
    env:
      JSB_PUBLIC_KEY: ${{ secrets.JSB_PUBLIC_KEY }}
      JSB_PRIVATE_KEY: ${{ secrets.JSB_PRIVATE_KEY }}
    steps:
      - uses: actions/checkout@v4

      - name: Install jsb
        run: |
          curl -sL https://github.com/jsonbankio/jsonbank-cli/releases/download/v0.1.1/jsb_0.1.1_linux_amd64.tar.gz | tar -xz jsb
          sudo mv jsb /usr/local/bin/

      - name: Push data to JSONBank
        run: jsb file update project/data.json data.json
```

## Development

```sh
make build   # build ./jsb
make test    # run tests
make install # symlink jsb into your PATH
```
