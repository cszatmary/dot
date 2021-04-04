# dot

`dot` is a small CLI for managing your dotfiles. It allows you to sync dotfile sources to the actual dotfiles.
This is useful if you wish to keep all your dotfiles in a git repopository since sources can be stored separately from the actual dotfiles.

## Installation

### Homebrew

```
brew install cszatmary/tap/dot
```

### Binary Release

You can manually download a binary release from [the release page](https://github.com/cszatmary/dot/releases)
or use a tool like [curl](https://curl.se/) or [wget](https://www.gnu.org/software/wget/).

### Install from source

Required Go version >= 1.16.

```
go get github.com/cszatmary/dot
```

## Usage

`dot` uses a registry to manage dotfiles. A registry is simple a directory with a `dot.yml` file and the dotfile sources.

First setup dot to use a registry:

```
dot setup -r <path to registry directory>
```

Now any time you want to update your dotfiles simply run:

```
dot apply
```

By default apply will update all dotfiles if no arguments are provided.
Arguments may be optionally provided to only update specific dotfiles.

```
dot apply vim zsh
```

### `dot.yml`

dot is configured using a `dot.yml` file which must be located in the root directory of a registry.

Ex:

```yml
dotfiles:
  git:
    src: git/gitconfig
    dst: ~/.gitconfig
  zsh:
    src: zsh/zshrc
    dst: ~/.zshrc
```

`dot.yml` must contain a top level `dotfiles` key which is a map of dotfile names to their configuration.
The name is used to identify the dotfile in the `apply` command.
`src` is the path to the source file in the registry and must be relative to the registry.
`dst` is the absolute path to the actual dotfile on your filesystem.

## License

dot is available under the [MIT License](LICENSE).

## Contributing

Contributions are welcome. Feel free to open an issue or submit a pull request.
