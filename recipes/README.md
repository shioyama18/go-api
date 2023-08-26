# Sample code taken from "Building Distributed Applications in Gin"

## Instructions

### Load data to MongoDB
```bash
$ docker run -it --rm --name mongodb \
    -v ~/mongo:/data/db \
    -v $PWD/recipes.json:/tmp/recipes.json \
    -e MONGO_INITDB_ROOT_USERNAME=admin \
    -e MONGO_INITDB_ROOT_PASSWORD=password \
    mongo:4.4.24 \
    mongoimport -d demo -c recipes --file /tmp/recipes.json --drop
```

### Running MongoDB
```bash
$ docker run --rm -d --name mongodb \
    -v ~/mongo:/data/db \
    -e MONGO_INITDB_ROOT_USERNAME=admin \
    -e MONGO_INITDB_ROOT_PASSWORD=password \
    -p 27017:27017 \
    mongo:4.4.24
```

### Running Redis
```bash
$ docker run --rm -d --name redis \
    -v $PWD/conf:/usr/local/etc/redis \
    -p 6379:6379 \
    redis:7.2 \
    redis-server /usr/local/etc/redis/redis.conf
```

### Running the code
```bash
$ export MONGO_URI="mongodb://admin:password@localhost:27017/test?authSource=admin"
$ export MONGO_DATABASE=demo
$ go run main.go
```

### Generating and running Swagger
```bash
$ swagger generate spec -o ./swagger.json
$ swagger serve -F swagger ./swagger.json
```

