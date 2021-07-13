package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego-component/ees"
)

//  export EGO_DEBUG=true && go run main.go --config=config.toml
func main() {
	ego.New().Invoker(espost).Run()
}

func espost() error {
	comp := ees.Load("es").Build()

	// Build the request body.
	var b strings.Builder
	b.WriteString(`{"title" : "`)
	b.WriteString("hello")
	b.WriteString(`"}`)
	req := esapi.IndexRequest{
		Index: "ego_logger",
		Body:  strings.NewReader(b.String()),
	}
	res, err := req.Do(context.Background(), comp.Client)
	fmt.Println(res)
	fmt.Println(err)
	return nil
}
