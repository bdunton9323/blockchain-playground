#!/bin/bash

# clean
rm contract/build/*

# compile the contracts
solc --allow-paths "$PWD/node_modules/@openzeppelin/" --abi contract/DeliveryContract.sol -o contract/build
solc --allow-paths "$PWD/node_modules/@openzeppelin/" --bin contract/DeliveryContract.sol -o contract/build

# build the Go bindings
abigen --abi contract/build/DeliveryContract.abi \
    --pkg contract \
    --type DeliveryContract \
    --out contract/deliverycontract.go \
    --bin contract/build/DeliveryContract.bin