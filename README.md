# EnvFile
A Go library for loading environment variables from files.

## Installation
As a library

```
go get github.com/afonichev/envfile
```

## Usage
Your .envfile file in the root of your project:

```
### Application configuration.

# secret key
export SECRET_KEY = qo~tVz>0|@(1'Gao>kTxZPeVsu`M5+QY2k4"v4c$^6cA,6NrNn9pf6GSyGyyXC@

# server port
overload PORT = 3000

### Database configuration.

# database connection address
export DB_URI = mongodb://{ DB_USERNAME }:{ DB_PASSWORD }@{ DB_HOST }:{ DB_PORT }/{ DB_NAME }

# database host
export DB_HOST = localhost

# database port
export DB_PORT = 27017

# database name
export DB_NAME = main

# username to access the database
export DB_USERNAME = admin

# user password to access the database
export DB_PASSWORD = qwerty
```

Your application:

```go
package main

import (
    "fmt"
    "os"

    "github.com/afonichev/envfile"
)

func main() {

    if err := envfile.Load(); err != nil {
        panic(err)
    }

    fmt.Println("PORT.......:", os.Getenv("PORT"))
    fmt.Println("SECRET_KEY.:", os.Getenv("SECRET_KEY"))
    fmt.Println("DB_URI.....:", os.Getenv("DB_URI"))
    fmt.Println("DB_HOST....:", os.Getenv("DB_HOST"))
    fmt.Println("DB_PORT....:", os.Getenv("DB_PORT"))
    fmt.Println("DB_NAME....:", os.Getenv("DB_NAME"))
    fmt.Println("DB_USERNAME:", os.Getenv("DB_USERNAME"))
    fmt.Println("DB_PASSWORD:", os.Getenv("DB_PASSWORD"))
}
```