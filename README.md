# go-api
Sample code taken from "Building Distributed Applications in Gin"

## Instructions

### Running the code
```bash
$ docker build -t api
$ docker-compose up -d
```

### Generating and running Swagger
```bash
$ swagger generate spec -o ./swagger.json
$ swagger serve -F swagger ./swagger.json
```