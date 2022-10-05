# blockchain-playground
This is an example application for me to experiment with blockchain and smart contracts.
A smart contract is a piece of code that runs in the blockchain that automatically
execute when certain conditions are met. For example, automatically release funds when
token holders vote. This is a useful tool in situations where the parties involved do
not trust each other.

In this demo, I define an Ethereum smart contract that simulates a shipment scenario.
A customer orders a product from the vendor. The vendor creates a non-fungible token (NFT)
from the delivery contract (which is viewable by all parties). To accept delivery, the
designated recipient must purchase the token from the vendor. This serves as definitive 
proof that the goods have traded hands, _and_ that the designated recipient was the one
who received them. 

## Applicability in the Real World
This is a toy scenario, but solutions like this have real-world value. The following scenarios
should be relatable:
- "I paid for expedited shipping, and it arrived too late to be useful"
- Customer: "I didn't get my stuff!". Vendor: "Too bad, my records show it was delivered." 
- "I asked for signature verification, but the driver just dropped it on my porch and left"

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
If you have an existing contract deployed and you don't want to recreate it, simply provide the address:
```
go run . -privatekey "..." -contractAddress "0xa8BBE18821035E7CBf64dA9d784e2846994b174E"
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

## Using the API
If you have the microservice running, you can view the interactive swagger page at http://localhost:8080/swagger/index.html.

If you prefer static docs over interactive ones, they can be found at [./docs](https://github.com/bdunton9323/blockchain-playground/tree/main/docs).

Here is some test data to get you going:
Server's data: 
- private key: `ae65abc8077ef5dd90eb22615f6ae708196bd4e580eae02a09d671cd83305c7b`
- address: `0x6066A53027eD103D934cD122Cd0C7AF2b9279c69`
- 
Customer's data: 
- private key: `e958f5d3e336803b8b23c389e77d6b29a74ff0d369f0a1d8aeeec1e27254624b`
- address: `0x7E0C39B48D52ADBc8660c1B03288Ef189787A133`

### Demonstration flow
These can all be done through the swagger UI or your tool of choice.
1. Place an order:
    This tells the microservice to create an order. You should see it in the database as well as the server's logs.
    This will mint a token and assign it to the server's account address.
    ```
    curl -X 'POST' \
        'http://localhost:8080/api/v1/order?itemId=7&buyerAddress=0x7E0C39B48D52ADBc8660c1B03288Ef189787A133' \
        -H 'accept: application/json' \
        -d ''
    ```
2. Grab the order ID from the response and use it to see who owns the token. This executes a method in the contract.
    ```
    curl -X 'GET' \
        'http://localhost:8080/api/v1/order/{orderId}' \
        -H 'accept: application/json'
    ```
3. Accept delivery of the order:
    This will cause the customer to purchase the token from the vendor for the total cost plus shipping.
    ```
    curl -X 'POST' \
        'http://localhost:8080/api/v1/order/c5127170-997e-4038-a2b2-a85af94a633c?customerKey=e958f5d3e336803b8b23c389e77d6b29a74ff0d369f0a1d8aeeec1e27254624b' \
        -H 'accept: application/json' \
        -H 'Content-Type: application/json' \
        -d '{"status": "delivered"}'
    ```
4. See that the customer now owns the token:

5. Burn the token
    The customer does not need to retain the delivery receipt forever, so destroy it!
    ```
    curl -X 'POST' \
        'http://localhost:8080/api/v1/order/c5127170-997e-4038-a2b2-a85af94a633c?customerKey=e958f5d3e336803b8b23c389e77d6b29a74ff0d369f0a1d8aeeec1e27254624b' \
        -H 'accept: application/json' \
        -H 'Content-Type: application/json' \
        -d '{"status": "canceled"}'
    ```

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