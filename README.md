# blockchain-playground
Example application for me to experiment with blockchain

## Running it
### Compile the contract and generate bindings
```
cd contract
solc --abi simplestorage.sol -o build
solc --bin simplestorage.sol -o build
abigen --abi build/SimpleStorage.abi --pkg contract --type Storage --out simplestorage.go --bin build/SimpleStorage.bin
```

### Run directly
```
go run .
```

### Build and run using docker
```
docker build -t blockchain-playground .
docker run --rm -p 3000:3000 blockchain-playground
```