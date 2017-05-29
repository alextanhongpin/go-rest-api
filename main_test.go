package main_test

import (
	"testing"
)

type MockJobService struct{}

func (svc MockJobService) fetchMany() ([]Job, error) {
	return []Job {
		Job{Name: "123"}
	}
}

func (svc MockJobService) fetchOne(id string) (Job, error) {
	return Job{ID: id, Name: "test"}
}

func (svc MockJobService) updateOne(id string, job Job) error {
	
}


func TestHttpGet(t *testing.T) {
	mockService := MockJobService{}
	getJobsHandler(mockService)
	r := http.NewRequest("GET", "/books", nil)
}
