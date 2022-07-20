package main

import (
	"Microservices_Go_Caching_with_Redis/controller"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	log.Println("starting server")

	http.HandleFunc("/api/", controller.NewAPI().RedisHandler)

	http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), nil)
	log.Println("Started the serer")
}
