#!/bin/bash

# clean
rm contract/build/*

# compile the contracts
solc --allow-paths "$PWD/node_modules/@openzeppelin/" --abi contract/DeliveryToken.sol -o contract/build
solc --allow-paths "$PWD/node_modules/@openzeppelin/" --bin contract/DeliveryToken.sol -o contract/build

# build the Go bindings
abigen --abi contract/build/DeliveryToken.abi \
    --pkg contract \
    --type DeliveryToken \
    --out contract/deliverytoken.go \
    --bin contract/build/DeliveryToken.bin