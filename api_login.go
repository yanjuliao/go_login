package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("postgres", "host=localhost port=5432 user=yan password=master dbname=BD01 sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = createUsersTable()
	if err != nil {
		log.Fatal(err)
	}

	err = createDefaultUser()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/login", handleLogin)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createUsersTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) NOT NULL,
			password VARCHAR(255) NOT NULL
		)
	`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func createDefaultUser() error {
	// Verifica se o usuário padrão já existe
	query := "SELECT COUNT(*) FROM users WHERE username = $1"
	var count int
	err := db.QueryRow(query, "usuario_padrao").Scan(&count)
	if err != nil {
		return err
	}

	// Se o usuário padrão não existe, insere-o na tabela
	if count == 0 {
		query = "INSERT INTO users (username, password) VALUES ($1, $2)"
		_, err = db.Exec(query, "usuario_padrao", "senha_padrao")
		if err != nil {
			return err
		}
	}

	return nil
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Consulta no banco de dados para verificar as credenciais do usuário
	query := "SELECT COUNT(*) FROM users WHERE username = $1 AND password = $2"
	var count int
	err = db.QueryRow(query, user.Username, user.Password).Scan(&count)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if count == 1 {
		response := Response{
			Success: true,
			Message: "Login efetuado com sucesso!",
		}
		json.NewEncoder(w).Encode(response)
	} else {
		response := Response{
			Success: false,
			Message: "username ou password inválidos",
		}
		json.NewEncoder(w).Encode(response)
	}
}
