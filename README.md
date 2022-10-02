# blockchain-playground
Example application for me to experiment with blockchain

## Running it
### Compile the contract and generate bindings
Install dependencies. From the root of the project:
```
npm install @openzeppelin/contracts@3.4.2
```

Build the contracts:
```
# compile the contracts
solc --allow-paths "$PWD/node_modules/@openzeppelin/" --abi contract/nft.sol -o contract/build
solc --allow-paths "$PWD/node_modules/@openzeppelin/" --bin contract/nft.sol -o contract/build

# build the Go bindings
abigen --abi contract/build/MyToken.abi \
    --pkg contract \
    --type MyToken \
    --out contract/mytoken.go \
    --bin contract/build/MyToken.bin
```

### Run the microservice
```
go run .
```

## Build and run using docker
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