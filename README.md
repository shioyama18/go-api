# go-api
Sample code taken from "Building Distributed Applications in Gin"

## Instructions

### Running the code
```bash
$ docker build -t api .
$ docker-compose up -d
```

### Running the test
```bash
$ cd cmd/server
$ docker-compose up -d
$ MONGO_URI="mongodb://admin:password@localhost:27017/test?authSource=admin&readPreference=primary&ssl=false" MONGO_DATABASE=demo REDIS_URI=localhost:6379 go test
```

### Generating and running Swagger
```bash
$ swagger generate spec -o ./swagger.json
$ swagger serve -F swagger ./swagger.json
```