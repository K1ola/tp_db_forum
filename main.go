package main

import (
	"github.com/K1ola/tp_db_forum/database"
	"fmt"
	"flag"
	"github.com/valyala/fasthttp"
	"github.com/K1ola/tp_db_forum/router"
)

const (
	port = ":5000"
	//dbURI = "postgresql://postgres:kate@localhost:5432/forum"
	dbURI = "postgresql://docker:docker@localhost:5432/docker"
)

func main() {
	err := database.Connect(dbURI)
	if err != nil {
		fmt.Println(err)
		return 
	}
	defer database.Disconn()

	r := router.NewRouter()
	
	flag.Parse()
	fmt.Println("Server started")
	
	err = fasthttp.ListenAndServe(port, r.Handler)
	if err != nil {
		fmt.Println(err.Error())
	}
}




