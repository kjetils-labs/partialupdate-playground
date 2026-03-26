package main

import (
	"partialupdate/internal/routes"
	v1 "partialupdate/internal/routes/v1"
)

func main() {
	router, err := routes.Setup("mongodb://admin:secret@localhost:27017", "person", "person", v1.SetupV1Routes)
	if err != nil {
		panic(err)
	}

	router.Start("localhost:8080")
}
