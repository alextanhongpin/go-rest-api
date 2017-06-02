package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/alextanhongpin/simple-api/Godeps/_workspace/src/github.com/go-sql-driver/mysql"
	"github.com/alextanhongpin/simple-api/Godeps/_workspace/src/github.com/julienschmidt/httprouter"
)

type (
	// Config contains the mapping from our config.json to golangs struct
	Config struct {
		// The data source name for the database. Required.
		DSN string `json:"dsn"`
	}

	// Env contains our app env
	Env struct {
		// The db context
		DB *sql.DB
		// The config
		Config Config
	}

	Job struct {
		ID        int64     `db:"id" json:"id"`
		Name      string    `db:"name" json:"name"`
		CreatedAt time.Time `db:"created_at" json:"created_at"`
	}
)

// Global env
var env Env

func main() {
	// Flags
	var (
		// go run main.go -port=4000 -config=config.json
		port = flag.Int("port", 8080, "The server port.")
		cfg  = flag.String("config", "config.json", "Path to a config file.")
	)
	flag.Parse()

	// Load the config from the json file from the path specified
	env.Config = setupConfig(*cfg)

	// Setup the connection to our db
	env.DB = setupDB(env.Config.DSN)
	defer env.DB.Close()

	// Setup router
	router := httprouter.New()

	// Setup routes
	router.GET("/api/jobs", getJobsHandler)
	router.GET("/api/jobs/:id", getJobHandler)
	router.POST("/api/jobs", createJobHandler)
	router.DELETE("/api/jobs/:id", deleteJobHandler)
	router.PUT("/api/jobs/:id", updateJobHandler)

	// Start server
	fmt.Printf("listening to port *:%d. press ctrl + c to cancel", *port)
	http.ListenAndServe(fmt.Sprintf(":%d", *port), router)
}

// Get a list of job
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

// Get a job based on the specified id
func getJobHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	var job Job
	err := env.DB.QueryRow("SELECT id, name, created_at FROM job WHERE id=?", id).Scan(&job.ID, &job.Name, &job.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Fprint(w, `{"data": null }`)
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
	id := ps.ByName("id")

	var job Job
	err := json.NewDecoder(r.Body).Decode(&job)

	_, err = env.DB.Exec("UPDATE job SET name=? WHERE id=?", job.Name, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func deleteJobHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	_, err := env.DB.Exec("DELETE FROM job WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
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
	w.WriteHeader(http.StatusCreated)
}

// setupConfig takes a pointer reference and set the config loaded
func setupConfig(path string) Config {
	var config Config

	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}

	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		log.Fatal(err)
	}

	return config
}

// setupDB prepares the db for connections
func setupDB(dataSourceName string) *sql.DB {
	var db *sql.DB

	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}

	// Limit to 150 connections, default is unlimited
	db.SetMaxOpenConns(150)

	// Ping to ensure connection is available
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	return db
}
