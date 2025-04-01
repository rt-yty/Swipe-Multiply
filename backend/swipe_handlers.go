package backend

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// Profile представляет данные профиля пользователя.
type Profile struct {
	Login       string `json:"login"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	City        string `json:"city"`
	BirthDate   string `json:"birthDate"`
	Gender      string `json:"gender"`
	Description string `json:"description"`
}

// DiscoverProfileHandler возвращает случайного пользователя, исключая того, кто ищет.
func DiscoverProfileHandler(w http.ResponseWriter, r *http.Request) {
	// Извлекаем логин пользователя, который ищет, из параметров URL
	searchLogin := r.URL.Query().Get("login")
	if searchLogin == "" {
		http.Error(w, "Параметр login обязателен", http.StatusBadRequest)
		return
	}

	// Выбираем случайного пользователя, чей login не равен searchLogin
	query := `
		SELECT login, first_name, last_name, city, birth_date, gender, description
		FROM user_profiles
		WHERE login != $1
		ORDER BY RANDOM()
		LIMIT 1;
	`

	var profile Profile
	err := Db.QueryRow(query, searchLogin).Scan(
		&profile.Login,
		&profile.FirstName,
		&profile.LastName,
		&profile.City,
		&profile.BirthDate,
		&profile.Gender,
		&profile.Description,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Нет других пользователей", http.StatusNotFound)
			return
		}
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}
