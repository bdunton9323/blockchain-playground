#!/bin/bash

# clean
rm contract/build/*

# compile the contracts
solc --allow-paths "$PWD/node_modules/@openzeppelin/" --abi contract/nft.sol -o contract/build
solc --allow-paths "$PWD/node_modules/@openzeppelin/" --bin contract/nft.sol -o contract/build

# build the Go bindings
abigen --abi contract/build/MyToken.abi \
    --pkg contract \
    --type MyToken \
    --out contract/mytoken.go \
    --bin contract/build/MyToken.bin