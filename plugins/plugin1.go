package main

import (
	"net/http"
	"fmt"
)

func Plugin(r http.Request)  {
	fmt.Println(r.UserAgent())
}