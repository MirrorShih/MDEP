package main

import (
	routes "MDEP/router"
)

func main() {
	router := routes.NewRouter()
	router.Run(":8000")
}
