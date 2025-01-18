package handlers

import (
	"fmt"
	"net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `<html><body><a href="/login">Login with Google</a></body></html>`)
}
