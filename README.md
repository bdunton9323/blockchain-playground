# blockchain-playground
Example application for me to experiment with blockchain.

This simulates a delivery tracking use case. 

TODO: say more

## Running it
### Compile the contract and generate bindings
Install dependencies. From the root of the project:
```
npm install @openzeppelin/contracts@3.4.2
```

Build the contracts and the Go bindings:
```
./rebuild_contracts.sh
```

### Run the microservice
```
go run .
```

## Build and run using docker
If you have docker compose installed as a plugin to the server version of docker:
```
docker compose up -d
```
If you have `docker-compose` as a standalone tool:
```
docker-compose up -d
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