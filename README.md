# blockchain-playground
Example application for me to experiment with blockchain.

This simulates a delivery tracking use case. 

TODO: say more

## Running it
### Install dependencies
From the root of the project:
```
npm install @openzeppelin/contracts@3.4.2
```

Install Quorum, a development ethereum blockchain that runs locally in docker:
1. Make a directory for the quorum installation:
    ```
    ~$ mkdir kaleido-io
    ```
2. Check out the git repository somewhere on your machine
    ```
    ~$ cd kaleido-io
    ~/kaleido-io$ git clone git@github.com:kaleido-io/quorum-tools.git
    ~/kaleido-io$ git clone git@github.com:kaleido-io/quorum.git
    ```
3. Run quorum
    ```
    ~/kaleido-io$ cd quorum-tools/examples
    # on your machine you might need "docker-compose" instead of "docker compose"
    ~/kaleido-io/quorum-tools/examples$ docker compose up
    ```
4. Run the database locally. From the `blockchain-playground` checkout directory:
   ```
   ~/blockchain-playground$ cd database
   ~/blockchain-playground/database$ docker compose up -d
   ```
   This will take a minute to come up. Run `docker ps` and Make sure you see `mariadb` in the list of running containers


### Building the project
`cd` to the project root.

Build the contracts and the Go bindings:
```
./rebuild_contracts.sh
```

### Run the microservice
```
go run . -privatekey "ae65abc8077ef5dd90eb22615f6ae708196bd4e580eae02a09d671cd83305c7b"
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
docker run --rm -p 8080:8080 blockchain-playground
```

Once it's running, you can view the API documentation and try out the APIs here: http://localhost:8080/swagger/index.html!

## Developing
The API docs are generated with `swaggo`: https://github.com/swaggo/swag

Regenerate the API docs:
```
swag init -g controllers/router.go
```

Connect to database:
Install `mysql-client` through your package manager, then connect:
```
mysql -u db_user --password=mysqlPassword --host 127.0.0.1
```