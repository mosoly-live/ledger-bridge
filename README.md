# Mosoly Ledger bridge <!-- omit in toc -->

- [Development](#development)
  - [Setting up ledger bridge cache DB](#setting-up-ledger-bridge-cache-db)
  - [Re-creating database](#re-creating-database)
- [Run the application](#run-the-application)
- [Lint & build](#lint--build)
- [Metrics and debug counters](#metrics-and-debug-counters)

## Development

In order to locally develop mosoly ledger bridge sync some of the pre-requisites needs to be setup:

- Golang 1.12 runtime for building and running the instance. Install one from [golang.org](https://golang.org)
- Docker. Follow the [Docker get started guide]((https://www.docker.com/get-started)) in order to install it

### Setting up ledger bridge cache DB

A Docker image will be running Postgres DB instance. In order to make sure that our DB is persisted between shutdowns we must create a folder for storing DB files.

```sh
mkdir -p $HOME/docker/volumes/postgres
```

After persisted storage is prepare we will run A Docker image and map the folder for the DB files

```sh
docker run --rm \
  --name pg-docker \
   -e POSTGRES_USER=mosoly \
   -e POSTGRES_DB=ledger_bridge_cache \
   -e POSTGRES_PASSWORD=qwerty123456 \
   -d -p 45433:5432 \
   -v "${HOME}/docker/volumes/postgres:/var/lib/postgresql/data" \
   postgres
```

By running this command you will:

1. spin a new Postgres instance
2. create a database called `ledger_bridge_cache`
3. create a user `mosoly` with password `qwerty123456`
4. map `/var/lib/postgresql/data` folder to your local persisted storage `$HOME/docker/volumes/postgres`
5. Expose port `45433`  to your local environment in order to access the DB

The following are default configuration on the `mosoly-ledger-bridge` and are a part of `config/config.go`. **NOTE!** default values should not be used in production environment and we strongly recommend to run a virtual machine for running a DB instance to minimize any potential latency issues.

After PostgreSQL is confirmed to be running connect to the instance and execute and create a schema that can be found in `db/schema.sql` file.

### Re-creating database

If you ever need to re-create the database you will need to execute `db/drop.sql` to wipe data and tables after that you can re-create schema from `db/schema.sql`

Before running the script agains the database, please enter most recent value for `latest_processed_block_number` in query `INSERT INTO ethereum_blockchain ...`. If running Ropsten testnet use the latest block number from [https://ropsten.etherscan.io/](https://ropsten.etherscan.io/). If running Mainnet use the latest block number [https://etherscan.io/](https://etherscan.io/)

## Run the application

In order to run use the following command

```sh
make -j4 && ./artifacts/mosoly-ledger-bridge \
  -service.env "local" \
  -noconsul \
  -db.host "localhost" \
  -db.name "ledger_bridge_cache" \
  -db.user "mosoly" \
  -db.pass "qwerty123456" \
  -db.port "45433" \
  -ethereum.json.rpc.url "https://ropsten.infura.io/" \
  -app.mosoly.ops.account "54068abd52c9592304c068c1101c2a5773b23b79406348403b462a4ea8d40636" \
  -ethereum.passport.factory.address "0x35Cb95Db8E6d56D1CF8D5877EB13e9EE74e457F2" \
  -app.mosoly.did.address "0xD0cC759A5380525CbDeB3C1Dd59de3a21A637176" \
  -app.mosoly.backend.url "http://localhost:9001/api" \
  -app.mosoly.backend.token "ZXlKaGJHY2lPaUpJVXpJMU5pSjkuZXlKemRXSWlPaUpCVUZCVlUwVlNJbjAudG54Zk0xTG5xOE9KTGU3STVLVHFCa0luTzBPQ1FyM0xfbGh4VlIwcmR4bw=="
```

## Lint & build

Sort imports:

```sh
goimports -w $(find . -type f -name '*.go' -not -path "./vendor/*")
```

To lint and build the application with current changes, do `make -j4 all`. It installs all required Go tools, lints and builds the code. It is also what runs inside CI job container and it's usually faster if you do it first on your machine before pushing.

However, your IDE should probably be also able to tell most of the important problems. Go builds very fast and thus it's possible to see build errors and warnings very quickly while working.

You can modify the arguments the way you see fit for the feature you're developing. Please see `config/config.go` to find out what each argument helps us with.

## Metrics and debug counters

Service exposes metrics and some debug counters via HTTP at /debug/vars in JSON format: [http://localhost:8087/debug/vars](http://localhost:8087/debug/vars)

Prometheus metrics endpoint: [http://localhost:8087/metrics](http://localhost:8087/metrics)
