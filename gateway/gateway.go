package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Gateway service is running"))
}

func ProxyToAuthService(w http.ResponseWriter, r *http.Request) {
	// Eliminar el prefijo "/auth" del RequestURI para redirigir correctamente
	uri := strings.TrimPrefix(r.RequestURI, "/auth")
	url := "http://auth:8001" + uri
	log.Println("Redirecting to:", url)

	proxyRequest, err := http.NewRequest(r.Method, url, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(proxyRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	w.Write(body)
}

func ProxyToTodoService(w http.ResponseWriter, r *http.Request) {
	// Eliminar el prefijo "/todos" del RequestURI para redirigir correctamente
	uri := strings.TrimPrefix(r.RequestURI, "/todos")
	url := "http://todo:8002" + uri
	log.Println("Redirecting to:", url)

	proxyRequest, err := http.NewRequest(r.Method, url, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(proxyRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	w.Write(body)
}

func main() {
	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/auth/", ProxyToAuthService)
	http.HandleFunc("/todos/", ProxyToTodoService)

	log.Println("Gateway service running on port 8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
