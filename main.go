package main

import "net/http"

func main() {
	static := http.Dir("web/dist")

	http.Handle("/", http.FileServer(static))
	server := http.Server{
		Addr: ":8000",
	}
	server.ListenAndServe()
}

