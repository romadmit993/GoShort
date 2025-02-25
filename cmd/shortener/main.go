package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

var urlStore = map[string]string{
	"EwHXdJfB": "https://practicum.yandex.ru/", // Пример записи
}

func handle(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-Type")
		if contentType != "text/plain" {
			http.Error(w, "Неверный Content-Type", http.StatusBadRequest)
			return
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Ошибка чтения тела запроса", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		originalURL := string(body)
		fmt.Println("Получен URL:", originalURL)
		shortURL := "http://localhost:8080/EwHXdJfB"
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, shortURL)
	} else if r.Method == http.MethodGet {

		id := strings.TrimPrefix(r.URL.Path, "/")
		originalURL, exists := urlStore[id]
		if !exists {
			http.Error(w, "Сокращённый URL не найден", http.StatusBadRequest)
			return
		}
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handle)

	fmt.Println("Сервер запущен на http://localhost:8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
