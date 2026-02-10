package storage

import (
	"encoding/json"

	"github.com/cockroachdb/pebble"
	"github.com/virgiliusnanamanek02/gocron-dist/internal/scheduler"
)

type Store struct {
	db *pebble.DB
}

func NewStore(dir string) (*Store, error) {
	// Membuka database di folder tertentu (sesuai nama node)
	db, err := pebble.Open(dir, &pebble.Options{})
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) SaveJob(j *scheduler.Job) error {
	data, _ := json.Marshal(j)
	return s.db.Set([]byte(j.ID), data, pebble.Sync)
}

func (s *Store) GetAllJobs() ([]*scheduler.Job, error) {
	var jobs []*scheduler.Job
	iter, _ := s.db.NewIter(nil)
	for iter.First(); iter.Valid(); iter.Next() {
		var j scheduler.Job
		json.Unmarshal(iter.Value(), &j)
		jobs = append(jobs, &j)
	}
	iter.Close()
	return jobs, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
