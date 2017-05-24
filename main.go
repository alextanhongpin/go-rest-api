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

type Service interface {
	fetchMany() ([]Job, error)
	fetchOne(string) (Job, error)
	create(Job) error
	update(string, Job) error
	delete(string) error
}

type JobService struct {
	DB *sql.DB
}

func (js JobService) fetchMany() ([]Job, error) {
	var jobs []Job
	rows, err := js.DB.Query("SELECT id, name, created_at FROM job")
	defer rows.Close()
	if err != nil {
		return jobs, err
	}

	for rows.Next() {
		var job Job
		err := rows.Scan(&job.ID, &job.Name, &job.CreatedAt)
		if err != nil {
			return jobs, err
		}
		jobs = append(jobs, job)
	}

	if err = rows.Err(); err != nil {
		return jobs, err
	}
	return jobs, nil
}

func (js JobService) fetchOne(id string) (Job, error) {
	var job Job
	err := env.DB.QueryRow("SELECT id, name, created_at FROM job WHERE id=?", id).Scan(&job.ID, &job.Name, &job.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return job, nil

		} else {
			return job, err
		}
	}
	return job, nil
}

func (js JobService) update(id string, job Job) error {
	_, err = env.DB.Exec("UPDATE job SET name=? WHERE id=?", job.Name, id)
	return err
}

func (js JobService) delete(id string) error {
	_, err := env.DB.Exec("DELETE FROM job WHERE id = ?", id)
	return err
}

func (js JobService) create(job Job) error {
	_, err = env.DB.Exec("INSERT INTO job (name) values(?)", job.Name)
	return err
}

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

	// Define a new service. Use dependency injection.
	jobsvc := JobService{DB: env.DB}

	// Setup routes
	router.GET("/api/jobs", getJobsHandler(jobsvc))
	router.GET("/api/jobs/:id", getJobHandler(jobSvc))
	router.POST("/api/jobs", createJobHandler(jobsvc))
	router.DELETE("/api/jobs/:id", deleteJobHandler(jobsvc))
	router.PUT("/api/jobs/:id", updateJobHandler(jobsvc))

	// Start server
	fmt.Printf("listening to port *:%d. press ctrl + c to cancel", *port)
	http.ListenAndServe(fmt.Sprintf(":%d", *port), router)
}

func getJobsHandler(svc Service) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		jobs, err := svc.fetchMany()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jobs)
	}
}

// Get a job based on the specified id
func getJobHandler(svc Service) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		id := ps.ByName("id")
		job, err := svc.fetchOne(id)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(job)
	}
}
func updateJobHandler(svc Service) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		id := ps.ByName("id")

		var job Job
		err := json.NewDecoder(r.Body).Decode(&job)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = svc.update(id, job)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	}
}
func deleteJobHandler(svc JobService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		id := ps.ByName("id")

		err := svc.delete(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	}
}

func createJobHandler(svc JobService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		var job Job
		err := json.NewDecoder(r.Body).Decode(&job)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = svc.create(job)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	}
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
