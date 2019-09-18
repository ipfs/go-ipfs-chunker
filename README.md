# go-btfs-chunker

> go-btfs-chunker implements data Splitters for go-btfs.

`go-btfs-chunker` provides the `Splitter` interface. IPFS splitters read data from a reader an create "chunks". These chunks are used to build the BTFS DAGs (Merkle Tree) and are the base unit to obtain the sums that BTFS uses to address content.

The package provides a `SizeSplitter` which creates chunks of equal size and it is used by default in most cases, and a `rabin` fingerprint chunker. This chunker will attempt to split data in a way that the resulting blocks are the same when the data has repetitive patterns, thus optimizing the resulting DAGs.

## Table of Contents

- [Install](#install)
- [Usage](#usage)
- [Contribute](#contribute)
- [License](#license)

## Install

`go-btfs-chunker` works like a regular Go module:

```
> go get github.com/TRON-US/go-btfs-chunker
```

It uses [Gx](https://github.com/whyrusleeping/gx) to manage dependencies. You can use `make all` to build it with the `gx` dependencies.

## Usage

```
import "github.com/TRON-US/go-btfs-chunker"
```

## Contribute

PRs accepted.

Small note: If editing the README, please conform to the [standard-readme](https://github.com/RichardLitt/standard-readme) specification.

## License

MIT © TRON-US.
