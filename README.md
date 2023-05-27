# API Object Store - Kong

## Authors

- [@miravtmehta](https://www.github.com/miravtmehta)
# Documentation
## Project Structure

* `db.go` - Initialize DB connections
* `handlers.go` - Handle HTTP requests and execute model function calls
* `main.go` - main package to initiate routes and logger
* `model.go` - defines products database model and executes CRUD operations
* `services.go` - mapper service methods with ServiceProvider Interface for scoped usage
* `utils.go` - utility functions for the applications


## Getting Started

This Makefile provides a set of commands to help with the building, running, and cleaning of the endor project. The project includes creating go executables, setting up a postgres container, and running the executables for Mac and Linux. Below are the available commands:

```
make help        Show help for each of the Makefile recipes.
make clean       Removes go executables and  postgres container
make build       Runs "clean" and "setup" to build infra and dependencies and create go executables
make setup       Pull latest postgres image and run as container @ port 5432
make run-mac     Run go executable for Mac
make run-linux   Run go executable for Linux
```

## Commands

### help

The `help` command shows help for each of the Makefile recipes.

### clean

The `clean` command removes the go executables and the postgres container.

### build

The `build` command runs `clean` + `setup`creates the go executables after setting up the postgres container.

### setup

The `setup` command pulls the latest postgres image and runs it as a container at port 5432.

### run-mac

The `run-mac` command runs the go executable for Mac that performs CRUD operations on object data present in main.go.

### run-linux

The `run-linux` command runs the go executable for Linux that performs CRUD operations on object data present in main.go.


## Build local setup

Run `build` to install all necessary infrastructure and dependent packages and run `make run-linux` or `make run-mac` based on OS hosted on http://localhost:8080



## API Reference

#### Get all services

```http
  GET http://localhost:8080/services
```


#### Get specific service details inclusive versions available


```http
  GET http://localhost:8080/services/<service-name>
  Example : http://localhost:8080/services/reporting
```

#### Create new service


```http
  POST http://localhost:8080/services/<service-name>
  
  Example:
  curl --location 'http://localhost:8080/services' \
--header 'Content-Type: application/json' \
--data '{
    "versions": [
        323,
        34,
        34
    ],
    "name": "gymclass",
    "description": "quibusdam quia rerum in accusamus. sit inventore rem. harum magnam id nisi officia ratione."
}'
```

#### Delete specific service 


```http
  DELETE http://localhost:8080/services/<service-name>
  Example : http://localhost:8080/services/security
```


## API Operations - filtering, sorting, pagination

#### Filtering - Get list of service(s) matching **_text_**
```http
  http://localhost:8080/services?name=<text>
  http://localhost:8080/services?name=monit
```

##### Sorting - Get list of service(s) based on parameter ; By default sorted based on created_at

###### Ascending order 
```http
  http://localhost:8080/services?sort=created_at // Default
  http://localhost:8080/services?sort=name
  http://localhost:8080/services?sort=id
```

###### Descending order
```http
  http://localhost:8080/services?sort=-created_at
  http://localhost:8080/services?sort=-name
  http://localhost:8080/services?sort=-id
```

#### Pagination - Using offset and limit

###### Default limit = 50
```http
  http://localhost:8080/services?limit=4&offset=4
```

#### Generate DUMMY PG data and store -- Specifically expose for testing purpose ONLY

```http
  POST http://localhost:8080/dump
``` 

#### Clean DUMMY PG data -- Specifically expose for testing purpose ONLY

```http
  DELETE http://localhost:8080/dump
``` 