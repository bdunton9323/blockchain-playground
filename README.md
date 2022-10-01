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
docker compose up -d
```
This is suitable for a dev environment because of the way the database credentials are managed.

Note: until I get the go code in docker compose, run it like this:
```
docker build -t blockchain-playground .
docker run --rm -p 3000:3000 blockchain-playground
```

If you have `mysql-client` installed, you can connect to the database like this:
```
mysql -u db_user --password=mysqlPassword --host 127.0.0.1
```