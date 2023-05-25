# go-ipfs-chunker

> go-ipfs-chunker implements data Splitters for go-ipfs.

[![](https://img.shields.io/badge/made%20by-Protocol%20Labs-blue.svg?style=flat-square)](http://ipn.io)
[![](https://img.shields.io/badge/project-IPFS-blue.svg?style=flat-square)](http://ipfs.io/)
[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)
[![GoDoc](https://godoc.org/github.com/ipfs/go-ipfs-chunker?status.svg)](https://godoc.org/github.com/ipfs/go-ipfs-chunker)
[![Build Status](https://travis-ci.org/ipfs/go-ipfs-chunker.svg?branch=master)](https://travis-ci.org/ipfs/go-ipfs-chunker)

## ❗ This repo is no longer maintained.
👉 We highly recommend switching to the maintained version at https://github.com/ipfs/boxo/tree/main/chunker.
🏎️ Good news!  There is [tooling and documentation](https://github.com/ipfs/boxo#migrating-to-boxo) to expedite a switch in your repo. 

⚠️ If you continue using this repo, please note that security fixes will not be provided (unless someone steps in to maintain it).

📚 Learn more, including how to take the maintainership mantle or ask questions, [here](https://github.com/ipfs/boxo/wiki/Copied-or-Migrated-Repos-FAQ).


## Summary

`go-ipfs-chunker` provides the `Splitter` interface. IPFS splitters read data from a reader an create "chunks". These chunks are used to build the ipfs DAGs (Merkle Tree) and are the base unit to obtain the sums that ipfs uses to address content.

The package provides a `SizeSplitter` which creates chunks of equal size and it is used by default in most cases, and a `rabin` fingerprint chunker. This chunker will attempt to split data in a way that the resulting blocks are the same when the data has repetitive patterns, thus optimizing the resulting DAGs.

## Table of Contents

- [Install](#install)
- [Usage](#usage)
- [License](#license)

## Install

`go-ipfs-chunker` works like a regular Go module:

```
> go get github.com/ipfs/go-ipfs-chunker
```

## Usage

```
import "github.com/ipfs/go-ipfs-chunker"
```

Check the [GoDoc documentation](https://godoc.org/github.com/ipfs/go-ipfs-chunker)

## License

MIT © Protocol Labs, Inc.
