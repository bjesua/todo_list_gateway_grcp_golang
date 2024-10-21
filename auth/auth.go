package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("my_secret_key")

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func dbConn() *sql.DB {
	dbDriver := "mysql"
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")

	// Verificación para asegurarse de que las variables de entorno no estén vacías
	if dbUser == "" || dbPass == "" || dbName == "" {
		log.Fatal("Faltan variables de entorno para la conexión a la base de datos")
	}

	dsn := dbUser + ":" + dbPass + "@tcp(db:3306)/" + dbName
	var db *sql.DB
	var err error

	for i := 0; i < 5; i++ {
		db, err = sql.Open(dbDriver, dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		log.Println("Intentando reconectar a la base de datos...")
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal("Error al conectar con la base de datos:", err)
	}

	return db
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Auth service is running"))
}

func Register(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Hash del password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	db := dbConn()
	defer db.Close()

	// Insertar el usuario en la base de datos
	insert, err := db.Prepare("INSERT INTO users(username, password) VALUES(?, ?)")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error preparando la consulta:", err)
		return
	}
	_, err = insert.Exec(user.Username, string(hashedPassword))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error insertando el usuario:", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func Login(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db := dbConn()
	defer db.Close()

	// Obtener el usuario de la base de datos
	var storedUser User
	err = db.QueryRow("SELECT id, username, password FROM users WHERE username=?", user.Username).Scan(&storedUser.ID, &storedUser.Username, &storedUser.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println("Usuario no encontrado o error:", err)
		return
	}

	// Comparar la contraseña
	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Generar el token JWT
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Retornar el token JWT en la respuesta JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}

func ListUsers(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	defer db.Close()

	rows, err := db.Query("SELECT id, username FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err = rows.Scan(&user.ID, &user.Username)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}

	// Retorna la lista de usuarios en formato JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func main() {
	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/register", Register)
	http.HandleFunc("/login", Login)
	http.HandleFunc("/users", ListUsers) // Nueva ruta para listar los usuarios

	log.Println("Auth service running on port 8001")
	log.Fatal(http.ListenAndServe(":8001", nil))
}
