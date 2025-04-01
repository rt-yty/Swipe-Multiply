package backend

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var Db *sql.DB

func InitDB() {
	// Подключение к системной базе данных PostgreSQL
	sysConnStr := "host=localhost port=5432 user=postgres password=123 dbname=postgres sslmode=disable"
	sysDB, err := sql.Open("postgres", sysConnStr)
	if err != nil {
		log.Fatal("Ошибка подключения к системной БД:", err)
	}
	defer sysDB.Close()

	if err = sysDB.Ping(); err != nil {
		log.Fatal("Ошибка подключения к системной БД:", err)
	}

	targetDBName := "tinder_clone"

	// Проверка существования целевой базы данных
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)"
	err = sysDB.QueryRow(checkQuery, targetDBName).Scan(&exists)
	if err != nil {
		log.Fatal("Ошибка проверки существования базы данных:", err)
	}

	if !exists {
		log.Printf("База данных %s не существует. Создаем...", targetDBName)
		_, err = sysDB.Exec(fmt.Sprintf("CREATE DATABASE %s", targetDBName))
		if err != nil {
			log.Fatal("Ошибка создания базы данных:", err)
		}
		log.Printf("База данных %s успешно создана.", targetDBName)
	} else {
		log.Printf("База данных %s уже существует.", targetDBName)
	}

	// Подключение к целевой базе данных
	targetConnStr := fmt.Sprintf("host=localhost port=5432 user=postgres password=123 dbname=%s sslmode=disable", targetDBName)
	Db, err = sql.Open("postgres", targetConnStr)
	if err != nil {
		log.Fatal("Ошибка подключения к целевой базе данных:", err)
	}
	if err = Db.Ping(); err != nil {
		log.Fatal("Ошибка подключения к целевой базе данных:", err)
	}

	// Создание таблицы users для регистрации и логина с использованием login как PRIMARY KEY
	createUsersTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		login TEXT PRIMARY KEY,
		password TEXT NOT NULL,
		registration_complete BOOLEAN NOT NULL DEFAULT false
	);
	`
	if _, err := Db.Exec(createUsersTableQuery); err != nil {
		log.Fatal("Ошибка создания таблицы users:", err)
	}

	// Создание таблицы user_profiles для хранения данных профиля, с использованием login и добавлением gender
	createProfilesTableQuery := `
	CREATE TABLE IF NOT EXISTS user_profiles (
		login TEXT PRIMARY KEY,
		first_name TEXT,
		last_name TEXT,
		city TEXT,
		birth_date DATE,
		gender TEXT,
		description TEXT,
		FOREIGN KEY (login) REFERENCES users(login) ON DELETE CASCADE
	);
	`
	if _, err := Db.Exec(createProfilesTableQuery); err != nil {
		log.Fatal("Ошибка создания таблицы user_profiles:", err)
	}
}
