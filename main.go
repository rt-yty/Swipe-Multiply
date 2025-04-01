package main

import (
	"AwfulTinder/backend" // убедитесь, что путь соответствует вашей структуре проекта
	"log"
	"net/http"
)

func main() {
	// Инициализация базы данных
	backend.InitDB()

	// Настройка маршрутов API
	http.HandleFunc("/register", backend.RegisterHandler)
	http.HandleFunc("/login", backend.LoginHandler)
	http.HandleFunc("/continue_registration", backend.ContinueRegistrationHandler)
	http.HandleFunc("/discover", backend.DiscoverProfileHandler) // новый обработчик

	// Обслуживание статических файлов
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)

	port := "8080"
	log.Printf("Сервер запущен на порту %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
