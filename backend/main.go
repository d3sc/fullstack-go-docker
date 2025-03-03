package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type User struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// main function
func main() {
	// connect database

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT, email TEXT)")

	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()

	router.HandleFunc("/api/go/users", getUsers(db)).Methods("GET")
	router.HandleFunc("/api/go/users", createUser(db)).Methods("POST")
	router.HandleFunc("/api/go/users/{id}", getUser(db)).Methods("GET")
	router.HandleFunc("/api/go/users/{id}", updateUser(db)).Methods("PUT")
	router.HandleFunc("/api/go/users/{id}", deleteUser(db)).Methods("DELETE")

	// set up middleware
	enhancedRouter := enableCORS(JsonContentTypeMiddleware(router))

	log.Fatal(http.ListenAndServe(":8000", enhancedRouter))
}

// create middleware func

func enableCORS(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// set CORS header
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// pass request middleware to all route
		next.ServeHTTP(w, r)
	})

}

func JsonContentTypeMiddleware(next http.Handler) http.Handler {
	// setup for json format
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// get all users
func getUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get rows of all user
		rows, err := db.Query("SELECT * FROM users")
		if err != nil {
			log.Fatal(err)
		}

		// defer rows.Close() digunakan untuk menutup koneksi database setelah fungsi selesai dieksekusi
		// ini penting untuk mencegah memory leak dan memastikan resource database dilepaskan dengan benar
		defer rows.Close()

		users := []User{} /* Create array of users */

		// check all user is available (no error)
		for rows.Next() {
			var user User
			// scan data per user
			// method Scan() digunakan untuk memindahkan nilai dari hasil query ke dalam variabel yang ditentukan
			// dalam kasus ini memindahkan nilai id, name, dan email dari hasil query ke dalam struct User
			if err := rows.Scan(&user.Id, &user.Name, &user.Email); err != nil {
				log.Fatal(err)
			}
			// append every single data of user in var users
			users = append(users, user)
		}
		// check err
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		// return json to web
		json.NewEncoder(w).Encode(users)
	}
}

// get user by id
func getUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get url params id
		vars := mux.Vars(r)
		id := vars["id"]

		// get 1 row user in database
		var user User
		// method Scan() digunakan untuk memindahkan nilai dari hasil query ke dalam variabel yang ditentukan
		// dalam kasus ini memindahkan nilai id, name, dan email dari hasil query ke dalam struct User
		err := db.QueryRow("SELECT * FROM users WHERE id = $1", id).Scan(&user.Id, &user.Name, &user.Email)

		// if data rows not found
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// return json to web
		json.NewEncoder(w).Encode(user)
	}
}

// create user
func createUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get request body from post
		var user User
		json.NewDecoder(r.Body).Decode(&user)

		// method Scan() digunakan untuk memindahkan nilai dari hasil query ke dalam variabel yang ditentukan
		// dalam kasus ini memindahkan nilai id, name, dan email dari hasil query ke dalam struct User
		err := db.QueryRow("INSERT INTO users (name, email) values ($1, $2) RETURNING id", user.Name, user.Email).Scan(&user.Id)

		// cehck err
		if err != nil {
			log.Fatal(err)
		}

		// return json
		json.NewEncoder(w).Encode(user)
	}
}

// update user
func updateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get url paramter id
		vars := mux.Vars(r)
		id := vars["id"]

		// get request body
		var user User
		json.NewDecoder(r.Body).Decode(&user)

		//* PENTING! (IMPORTANT!)
		//* Gunakan tool yang tepat untuk pekerjaan yang tepat
		//* Query() untuk SELECT dengan banyak baris
		//* QueryRow() untuk SELECT satu baris atau operasi dengan RETURNING
		//* Exec() untuk operasi non-SELECT tanpa RETURNING

		// exec update user
		_, err := db.Exec("UPDATE users SET name = $1, email = $2 WHERE id = $3 RETURNING id", user.Name, user.Email, id)
		if err != nil {
			log.Fatal(err)
		}

		// query check for updated user are exist?
		var updatedUser User
		err = db.QueryRow("SELECT * FROM users WHERE id = $1", id).Scan(&updatedUser.Id, &updatedUser.Name, &updatedUser.Email)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(updatedUser)

	}
}

// delete user
func deleteUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		// Gunakan Exec() karena tidak memerlukan data yang dikembalikan
		result, err := db.Exec("DELETE FROM users WHERE id = $1", id)
		if err != nil {
			log.Fatal(err)
		}

		// Periksa apakah ada baris yang terpengaruh
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Fatal(err)
		}

		if rowsAffected == 0 {
			// Jika tidak ada baris yang dihapus, berarti user tidak ditemukan
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "User tidak ditemukan"})
			return
		}

		// Kirim respons sukses
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "User berhasil dihapus"})
	}
}
