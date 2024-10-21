package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Todo struct {
	ID     int    `json:"id"`
	User   string `json:"user"`
	Task   string `json:"task"`
	Status string `json:"status"`
}

func dbConn() *sql.DB {
	dbDriver := "mysql"
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dsn := dbUser + ":" + dbPass + "@tcp(" + dbHost + ":3306)/" + dbName

	var db *sql.DB
	var err error

	// Intentos de conexión a la base de datos
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
	w.Write([]byte("TODO service is running"))
}

// CreateTodo permite insertar una nueva tarea en la tabla "todos"
func CreateTodo(w http.ResponseWriter, r *http.Request) {
	var todo Todo
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar el cuerpo de la solicitud"))
		return
	}

	// Asegurarse de que los campos requeridos estén presentes
	if todo.User == "" || todo.Task == "" || todo.Status == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Faltan campos requeridos"))
		return
	}

	db := dbConn()
	defer db.Close()

	insert, err := db.Prepare("INSERT INTO todos(user, task, status) VALUES(?, ?, ?)")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err)
		return
	}

	_, err = insert.Exec(todo.User, todo.Task, todo.Status)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Tarea creada exitosamente"))
}

// ListTodos permite listar todas las tareas en la tabla "todos"
func ListTodos(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	defer db.Close()

	rows, err := db.Query("SELECT id, user, task, status FROM todos")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		err = rows.Scan(&todo.ID, &todo.User, &todo.Task, &todo.Status)
		if err != nil {
			log.Fatal(err)
		}
		todos = append(todos, todo)
	}

	// Devuelve la lista de tareas en formato JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

// DeleteTodo permite eliminar una tarea de la tabla "todos"
func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	// Obtener el parámetro "id" de la URL
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Falta el parámetro id"))
		return
	}

	// Convertir el ID a entero
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("El parámetro id debe ser un número"))
		return
	}

	db := dbConn()
	defer db.Close()

	// Eliminar la tarea de la base de datos
	deleteStmt, err := db.Prepare("DELETE FROM todos WHERE id=?")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err)
		return
	}

	_, err = deleteStmt.Exec(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Tarea eliminada exitosamente"))
}

func main() {
	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/todos", CreateTodo)        // Método POST para crear una tarea
	http.HandleFunc("/todos/list", ListTodos)    // Nueva ruta GET para listar todas las tareas
	http.HandleFunc("/todos/delete", DeleteTodo) // Nueva ruta DELETE para eliminar tareas

	log.Println("TODO service running on port 8002")
	log.Fatal(http.ListenAndServe(":8002", nil))
}
