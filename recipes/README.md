# Sample code taken from "Building Distributed Applications in Gin"

## Instructions

### Starting MongoDB
```bash
$ docker run -d --name mongodb \
    -v ~/mongo:/data/db \
    -e MONGO_INITDB_ROOT_USERNAME=admin \
    -e MONGO_INITDB_ROOT_PASSWORD=password \
    -p 27017:27017 \
    mongo:4.4.24
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

