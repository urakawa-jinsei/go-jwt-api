package main

import (
	"fmt"
	"net/http"

	"go-jwt-api/internal/handlers"
	"go-jwt-api/internal/middleware"

	"github.com/gorilla/mux"
)

func Protected(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	user := r.Context().Value("username").(string)
	fmt.Fprintf(w, "Hello, %s! This is a protected endpoint.", user)
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/login", handlers.Login).Methods("POST")
	r.Handle("/protected",
		middleware.JwtMiddleware(http.HandlerFunc(Protected)),
	).Methods("GET")

	http.ListenAndServe(":8000", r)
}
