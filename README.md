# react-rm

This program serves as a small command-line utility to remove all reactions 
from any issue in a given repository.

## Prerequisites

### Github API Token

This program uses the Github personal access token for authorization and authentication.

To create a new Github Token, follow the [GitHub documentation](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token#creating-a-token).

This application requires the `repo` access

![Image of repository settings](./docs/repo.png)

### Configuration file

The program expects a configuration file. See the example [github-cli](./docs/github-cli.yml).

Copy this configuration file into the default `~/.config` location:

```
cp docs/github-cli.yml ~/.config/
```

Alternatively, you can use any other location on the system. The program uses the `-c` flag for
passing a configuration file path, e.g.:

```
react-rm -c github-cli
```

Edit the configuration file with your values. 

## Building

To build the Go binary, execute:

```
make
```

To cross-compile for Linux and macOS, execute:

```
make build-all
```

Binaries are saved into the `bin` directory

## Executing

To execute the program, execute: 

```
make run
```