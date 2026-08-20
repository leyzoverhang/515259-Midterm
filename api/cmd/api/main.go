package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wongnok/internal/platform/database"
	"wongnok/internal/user"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, sqldb, err := database.Open(ctx, "postgresql://postgres:Pe@devp00l@localhost:5432?database=wongnok")
	if err != nil {
		log.Fatal("database connection:", err)
	}
	defer sqldb.Close()

	// Dependency injection
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	// Register path
	// curl -X GET http://localhost:8080/users/{id}
	http.HandleFunc("GET /users/{id}", userHandler.GetUser)

	// curl -X POST http://localhost:8080/users -H "Content-Type: application/json" -d '{"email":"taro@devpool.pea"}'
	http.HandleFunc("POST /users", userHandler.CreateUser)

	// Server
	serv := &http.Server{Addr: ":8080"}

	go serv.ListenAndServe()
	log.Println("server started at :8080")

	// Graceful shutdown
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), (10 * time.Second))
	defer cancel()

	serv.Shutdown(shutdownCtx)
}
