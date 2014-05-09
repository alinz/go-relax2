## Go Relax 2

Go Relax 2, is a second attempt to make a cleaner API for writing proper REST API.

### Usage

First install Go Relax 2 by running the following command

```
go get github.com/alinz/go-relax2
```

Create a new project and create a RelaxServer and start registering your handlers.

```go
package main

import (
  "github.com/alinz/go-relax2"
)

func main() {
  relaxServer := relax.NewRelax()

  relaxServer.RegisterHandler("GET", "/hello/{name}:string/welcome", func (req relax.RelaxRequest, res relax.RelaxResponse) {
      name := req.Param("name")
      res.Send("Hello "+name)
  })

  relaxServer.Listen("", 9000)
}
```
