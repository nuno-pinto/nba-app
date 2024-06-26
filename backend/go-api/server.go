package main

import (
	"os"

	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	serverport string
	db         *sql.DB
)

func init() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file, setting default port to 9000")
		serverport = "9000"
		return
	}

	serverport = os.Getenv("BACKEND_PORT")

	if serverport == "" {
		log.Println("BACKEND_PORT not set in .env file, setting default port to 9000")
		serverport = "9000"
	}
}

func main() {

	log.Println("Connecting to database")

	var err error
	db, err = OpenDB()
	if err != nil {
		log.Fatal("Failed connecting to database:", err)
	}
	defer db.Close()

	err = checkDB(db)
	if err != nil {
		log.Fatal("Error while checking database:", err)
	}

	server := getServer()

	log.Println("Starting server on port", serverport)
	err = server.ListenAndServe()

	if err != nil {
		log.Fatal("Failed to start server:", err)
	}

	defer server.Close()

	log.Println("Shutting down server")
}

func getServer() *http.Server {
	router := http.NewServeMux()
	router.HandleFunc("GET /hello", corsMiddleware(handleHello))
	router.HandleFunc("GET /player", corsMiddleware(handleGetPlayers))
	router.HandleFunc("GET /player/{id}", corsMiddleware(handleGetPlayerByID))
	router.HandleFunc("GET /random", corsMiddleware(handleGetRandomPlayer))

	portString := ":" + serverport

	server := http.Server{
		Addr:    portString,
		Handler: router,
	}

	return &server
}

func corsMiddleware(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,DELETE,PATCH,POST,PUT,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		handler(w, r)
	}
}

func handleHello(w http.ResponseWriter, r *http.Request) {
	log.Println("/hello request from", r.RemoteAddr)
	w.Write([]byte("Hello World"))
}

func handleGetPlayers(w http.ResponseWriter, r *http.Request) {
	log.Println("/players request from", r.RemoteAddr)

	players, err := GetAllPlayers(db)
	if err != nil {
		log.Println("Failed to get players:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(players)
}

func handleGetPlayerByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	log.Printf("/player/%s request from %s\n", idStr, r.RemoteAddr)

	id, err := strconv.Atoi(idStr)

	if err != nil {
		log.Println("Bad request, invalid ID:", idStr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	player, err := GetPlayerById(db, id)
	if err != nil {
		log.Println("Failed to get player by ID:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}

func handleGetRandomPlayer(w http.ResponseWriter, r *http.Request) {
	log.Println("/random request from", r.RemoteAddr)

	player, err := GetRandomPlayer(db)
	if err != nil {
		log.Println("Failed to get random player:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}

func checkDB(db *sql.DB) error {

	playerCount, err := CountPlayers(db)
	if err != nil {
		return err
	}

	if playerCount == 0 {
		log.Println("Adding players to database")

		players, err := GetPlayerData()
		if err != nil {
			return err
		}

		err = AddAllPlayers(db, players)
		if err != nil {
			return err
		}
	}

	return nil
}
