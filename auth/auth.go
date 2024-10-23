package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"

	pb "auth/auth/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

var jwtKey = []byte("my_secret_key")

type authServiceServer struct {
	pb.UnimplementedAuthServiceServer
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// Nueva estructura para manejar la contraseña
type UserWithPassword struct {
	ID       int32
	Username string
	Password string
}

func dbConn() *sql.DB {
	dbDriver := "mysql"
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")

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

// Implementación del método Register
func (s *authServiceServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	db := dbConn()
	defer db.Close()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	insert, err := db.Prepare("INSERT INTO users(username, password) VALUES(?, ?)")
	if err != nil {
		return nil, err
	}
	_, err = insert.Exec(req.Username, string(hashedPassword))
	if err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{Message: "Usuario registrado exitosamente"}, nil
}

func (s *authServiceServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	db := dbConn()
	defer db.Close()

	var storedUser UserWithPassword
	err := db.QueryRow("SELECT id, username, password FROM users WHERE username=?", req.Username).Scan(&storedUser.ID, &storedUser.Username, &storedUser.Password)
	if err != nil {
		return nil, err
	}

	// Comparar la contraseña
	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(req.Password))
	if err != nil {
		return nil, err
	}

	// Generar el token JWT
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Username: req.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return nil, err
	}

	return &pb.LoginResponse{Token: tokenString}, nil
}

// Implementación del método ListUsers
func (s *authServiceServer) ListUsers(ctx context.Context, req *emptypb.Empty) (*pb.ListUsersResponse, error) {
	db := dbConn()
	defer db.Close()

	rows, err := db.Query("SELECT id, username FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*pb.User
	for rows.Next() {
		var user pb.User
		err = rows.Scan(&user.Id, &user.Username)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return &pb.ListUsersResponse{Users: users}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Registrar el servicio AuthServiceServer en el servidor gRPC
	pb.RegisterAuthServiceServer(grpcServer, &authServiceServer{})

	// Habilitar la reflexión para permitir que herramientas como grpcurl descubran los servicios
	reflection.Register(grpcServer)

	log.Println("gRPC server is running on port 50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
