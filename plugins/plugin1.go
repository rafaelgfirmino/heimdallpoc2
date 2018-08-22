package main

import (
	"fmt"
	"net/http"
)

func Plugin(r http.Request) {
	fmt.Println(r.UserAgent())
}
