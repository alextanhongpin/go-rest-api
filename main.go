package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
)

type Configuration struct {
	Port       int    `json:"port"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBDatabase string `json:"db_database"`
}
type Enviroment struct {
	DB     *sql.DB
	Config Configuration
}

type Job struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

var env Enviroment

func main() {

	// Setup Config
	pathToConfig := os.Getenv("CONFIG")
	os.Unsetenv("CONFIG")
	if pathToConfig == "" {
		pathToConfig = "config.json"
	}
	loadConfig(pathToConfig)

	// Setup database
	setupDB()

	// Setup router
	router := httprouter.New()

	// Setup routes
	router.GET("/api/jobs", getJobsHandler)
	router.GET("/api/jobs/:id", getJobHandler)
	router.POST("/api/jobs", createJobHandler)
	router.DELETE("/api/jobs/:id", deleteJobHandler)
	router.PUT("/api/jobs/:id", updateJobHandler)

	// Start server
	fmt.Printf("listening to port *:%d. press ctrl + c to cancel", env.Config.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", env.Config.Port), router)
}

func getJobsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	rows, err := env.DB.Query("SELECT id, name, created_at FROM job")
	defer rows.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var jobs []Job
	for rows.Next() {
		var job Job
		err := rows.Scan(&job.ID, &job.Name, &job.CreatedAt)
		if err != nil {

			http.Error(w, err.Error(), http.StatusBadRequest)
			return

		}
		jobs = append(jobs, job)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func getJobHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	idStr := ps.ByName("id")
	idInt64, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var job Job
	err = env.DB.QueryRow("SELECT id, name, created_at FROM job WHERE id=?", idInt64).Scan(&job.ID, &job.Name, &job.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func updateJobHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	idStr := ps.ByName("id")
	idInt64, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var job Job
	err = json.NewDecoder(r.Body).Decode(&job)

	_, err = env.DB.Exec("UPDATE job SET name=? WHERE id=?", job.Name, idInt64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return

	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	// json.NewEncoder(w).Encode(job)
	w.Write([]byte(""))
}

func deleteJobHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	idStr := ps.ByName("id")
	idInt64, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = env.DB.Exec("DELETE FROM job WHERE id = ?", idInt64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte(""))
}
func createJobHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var job Job
	err := json.NewDecoder(r.Body).Decode(&job)

	_, err = env.DB.Exec("INSERT INTO job (name) values(?)", job.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte(""))
}

// loadConfig reads the config for a json file
func loadConfig(path string) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}

	err = json.NewDecoder(file).Decode(&env.Config)
	if err != nil {
		log.Fatal(err)
	}
}

// setupDB prepares the db for connections
func setupDB() {
	var err error
	user := env.Config.DBUser
	password := env.Config.DBPassword
	dbName := env.Config.DBDatabase
	dataSourceName := fmt.Sprintf("%s:%s@/%s?parseTime=true", user, password, dbName)
	env.DB, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	// Limit to 150 connections, default is unlimited
	env.DB.SetMaxOpenConns(150)

	// Ping to ensure connection is available
	err = env.DB.Ping()
	if err != nil {
		log.Fatal(err)
	}
}
