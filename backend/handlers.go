package backend

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

// RegisterHandler обрабатывает регистрацию пользователя.
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	type RegistrationRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var req RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Не указано имя или пароль", http.StatusBadRequest)
		return
	}

	// Проверка существования пользователя по login
	var exists bool
	err := Db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE login=$1)", req.Username).Scan(&exists)
	if err != nil {
		log.Printf("Ошибка проверки существования пользователя: %v", err)
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Пользователь с таким именем уже существует", http.StatusBadRequest)
		return
	}

	// Вставляем нового пользователя, registration_complete по умолчанию false.
	_, err = Db.Exec("INSERT INTO users (login, password) VALUES ($1, $2)", req.Username, req.Password)
	if err != nil {
		log.Println("Ошибка при записи в базу данных:", err)
		http.Error(w, "Ошибка записи в базу данных", http.StatusInternalServerError)
		return
	}

	log.Printf("Пользователь %s успешно зарегистрирован", req.Username)
	// Возвращаем JSON с флагом incomplete_registration, который равен true.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":                 "Регистрация прошла успешно. Требуется дополнительная регистрация.",
		"incomplete_registration": true,
	})
}

// LoginHandler обрабатывает вход пользователя.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Не указано имя или пароль", http.StatusBadRequest)
		return
	}

	var storedPassword string
	err := Db.QueryRow("SELECT password FROM users WHERE login=$1", req.Username).Scan(&storedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Пользователь не найден", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	if req.Password != storedPassword {
		http.Error(w, "Неверный пароль", http.StatusUnauthorized)
		return
	}

	// Проверка флага завершения регистрации.
	var registrationComplete bool
	err = Db.QueryRow("SELECT registration_complete FROM users WHERE login=$1", req.Username).Scan(&registrationComplete)
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	log.Printf("Пользователь %s успешно вошёл", req.Username)
	w.Header().Set("Content-Type", "application/json")
	// Всегда передаем флаг incomplete_registration, равный true если регистрация не завершена, иначе false.
	if !registrationComplete {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":                 "Вход выполнен успешно, но требуется дополнительная регистрация.",
			"incomplete_registration": true,
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":                 "Вход выполнен успешно.",
			"incomplete_registration": false,
		})
	}
}

// ContinueRegistrationHandler обрабатывает дополнительную регистрацию и заполнение профиля.
func ContinueRegistrationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	type AdditionalRegistration struct {
		Login       string `json:"login"` // Для идентификации пользователя
		FirstName   string `json:"firstName"`
		LastName    string `json:"lastName"`
		City        string `json:"city"`
		BirthDate   string `json:"birthDate"` // Формат: YYYY-MM-DD
		Gender      string `json:"gender"`    // Пол
		Description string `json:"description"`
	}
	var req AdditionalRegistration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	if req.Login == "" {
		http.Error(w, "Не указан логин", http.StatusBadRequest)
		return
	}

	// Проверяем, существует ли пользователь с данным login.
	var dummy string
	err := Db.QueryRow("SELECT login FROM users WHERE login=$1", req.Login).Scan(&dummy)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Пользователь не найден", http.StatusNotFound)
			return
		}
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	// Вставляем или обновляем профиль пользователя.
	query := `
	INSERT INTO user_profiles (login, first_name, last_name, city, birth_date, gender, description)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (login) DO UPDATE SET 
		first_name = EXCLUDED.first_name,
		last_name = EXCLUDED.last_name,
		city = EXCLUDED.city,
		birth_date = EXCLUDED.birth_date,
		gender = EXCLUDED.gender,
		description = EXCLUDED.description;
	`
	if _, err := Db.Exec(query, req.Login, req.FirstName, req.LastName, req.City, req.BirthDate, req.Gender, req.Description); err != nil {
		http.Error(w, "Ошибка записи профиля", http.StatusInternalServerError)
		return
	}

	// Обновляем флаг завершения регистрации.
	if _, err := Db.Exec("UPDATE users SET registration_complete=true WHERE login=$1", req.Login); err != nil {
		http.Error(w, "Ошибка обновления статуса регистрации", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Дополнительная информация успешно сохранена",
	})
}
