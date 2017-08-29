package main

import (
	"time"

	"github.com/restream/reindexer"
	_ "github.com/restream/reindexer/bindings/builtin"
)

const usersNamespace = "users"
const locationsNamespace = "locations"
const visitsNamespace = "visits"

type ReindexerStore struct {
	db *reindexer.Reindexer
}

func NewReindexerStore() (*ReindexerStore, error) {
	db := reindexer.NewReindex("builtin")
	if err := db.NewNamespace(usersNamespace, reindexer.DefaultNamespaceOptions(), &User{}); err != nil {
		return nil, err
	}
	if err := db.NewNamespace(locationsNamespace, reindexer.DefaultNamespaceOptions(), &Location{}); err != nil {
		return nil, err
	}
	if err := db.NewNamespace(visitsNamespace, reindexer.DefaultNamespaceOptions(), &Visit{}); err != nil {
		return nil, err
	}
	return &ReindexerStore{
		db: db,
	}, nil
}

func (s *ReindexerStore) isItemExists(ns string, id uint) bool {
	_, found := s.db.Query(ns).
		WhereInt("id", reindexer.EQ, int(id)).
		Get()
	return found
}

// User methods
func (s *ReindexerStore) CreateUser(u *User) error {
	if u.ID == 0 {
		return ErrMissingID
	}
	if s.isItemExists(usersNamespace, u.ID) {
		return ErrDup
	}
	if s.isEmailExists(u.Email, 0) {
		return ErrDup
	}
	return s.db.Upsert(usersNamespace, u)
}

func (s *ReindexerStore) CreateUsers(us []User) error {
	tx, err := s.db.BeginTx(usersNamespace)
	if err != nil {
		return err
	}
	for _, u := range us {
		if err := tx.Upsert(&u); err != nil {
			tx.Rollback()
			return err
		}
	}
	ts := time.Now()
	return tx.Commit(&ts)
}

func (s *ReindexerStore) UpdateUser(id uint, u *User) error {
	if id != u.ID {
		return ErrUpdateID
	}
	if !s.isItemExists(usersNamespace, id) {
		return ErrNotFound
	}
	if s.isEmailExists(u.Email, id) {
		return ErrDup
	}
	return s.db.Upsert(usersNamespace, u)
}

func (s *ReindexerStore) isEmailExists(email string, id uint) bool {
	elem, found := s.db.Query(usersNamespace).
		WhereString("email", reindexer.EQ, email).
		Get()
	if !found {
		return false
	}
	u := elem.(*User)
	return u.ID != id
}

func (s *ReindexerStore) GetUser(id uint, u *User) error {
	elem, found := s.db.Query(usersNamespace).
		WhereInt("id", reindexer.EQ, int(id)).
		Get()
	if !found {
		return ErrNotFound
	}
	*u = *elem.(*User)
	return nil
}

func (s *ReindexerStore) GetUserVisits(id uint, q *UserVisitsQuery, visits *[]UserVisit) error {
	return nil
}

// Location methods
func (s *ReindexerStore) CreateLocation(l *Location) error {
	if l.ID == 0 {
		return ErrMissingID
	}
	if s.isItemExists(locationsNamespace, l.ID) {
		return ErrDup
	}
	return s.db.Upsert(locationsNamespace, l)
}

func (s *ReindexerStore) CreateLocations(ls []Location) error {
	tx, err := s.db.BeginTx(locationsNamespace)
	if err != nil {
		return err
	}
	for _, l := range ls {
		if err := tx.Upsert(&l); err != nil {
			tx.Rollback()
			return err
		}
	}
	ts := time.Now()
	return tx.Commit(&ts)
}

func (s *ReindexerStore) UpdateLocation(id uint, l *Location) error {
	if id != l.ID {
		return ErrUpdateID
	}
	if !s.isItemExists(locationsNamespace, id) {
		return ErrNotFound
	}
	return s.db.Upsert(locationsNamespace, l)
}

func (s *ReindexerStore) GetLocation(id uint, l *Location) error {
	elem, found := s.db.Query(locationsNamespace).
		WhereInt("id", reindexer.EQ, int(id)).
		Get()
	if !found {
		return ErrNotFound
	}
	*l = *elem.(*Location)
	return nil
}

func (s *ReindexerStore) GetLocationAvg(id uint, q *LocationAvgQuery) (float64, error) {
	return 0, nil
}

// Visit methods
func (s *ReindexerStore) CreateVisit(v *Visit) error {
	if v.ID == 0 {
		return ErrMissingID
	}
	if s.isItemExists(visitsNamespace, v.ID) {
		return ErrDup
	}
	return s.db.Upsert(visitsNamespace, v)
}

func (s *ReindexerStore) CreateVisits(vs []Visit) error {
	tx, err := s.db.BeginTx(visitsNamespace)
	if err != nil {
		return err
	}
	for _, v := range vs {
		if err := tx.Upsert(&v); err != nil {
			tx.Rollback()
			return err
		}
	}
	ts := time.Now()
	return tx.Commit(&ts)
}

func (s *ReindexerStore) UpdateVisit(id uint, v *Visit) error {
	if id != v.ID {
		return ErrUpdateID
	}
	if !s.isItemExists(visitsNamespace, id) {
		return ErrNotFound
	}
	return s.db.Upsert(visitsNamespace, v)
}

func (s *ReindexerStore) GetVisit(id uint, v *Visit) error {
	elem, found := s.db.Query(visitsNamespace).
		WhereInt("id", reindexer.EQ, int(id)).
		Get()
	if !found {
		return ErrNotFound
	}
	*v = *elem.(*Visit)
	return nil
}

func (s *ReindexerStore) Clear() error {
	return nil
}
