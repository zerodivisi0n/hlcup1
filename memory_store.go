package main

import (
	"sync"

	"github.com/emirpasic/gods/trees/redblacktree"
)

type MemoryStore struct {
	mu               sync.RWMutex
	users            map[uint]User
	locations        map[uint]Location
	visits           map[uint]Visit
	emails           map[string]uint
	visitsByUser     map[uint]*redblacktree.Tree
	visitsByLocation map[uint]*redblacktree.Tree
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:            make(map[uint]User, 10000),
		locations:        make(map[uint]Location, 10000),
		visits:           make(map[uint]Visit, 10000),
		emails:           make(map[string]uint, 10000),
		visitsByUser:     make(map[uint]*redblacktree.Tree, 10000),
		visitsByLocation: make(map[uint]*redblacktree.Tree, 10000),
	}
}

// User methods
func (s *MemoryStore) CreateUser(u *User) error {
	s.mu.Lock()
	err := s.createUser(u)
	s.mu.Unlock()
	return err
}

func (s *MemoryStore) CreateUsers(us []User) error {
	s.mu.Lock()
	var err error
	for _, u := range us {
		err = s.createUser(&u)
		if err != nil {
			break
		}
	}
	s.mu.Unlock()
	return err
}

func (s *MemoryStore) createUser(u *User) error {
	// called with acquired mu lock
	if u.ID == 0 {
		return ErrMissingID
	}
	if _, exists := s.users[u.ID]; exists {
		return ErrDup
	}
	if _, exists := s.emails[u.Email]; exists {
		return ErrDup
	}
	s.users[u.ID] = *u
	s.visitsByUser[u.ID] = redblacktree.NewWith(timestampComparator)
	return nil
}

func (s *MemoryStore) UpdateUser(id uint, u *User) error {
	s.mu.Lock()
	err := s.updateUser(id, u)
	s.mu.Unlock()
	return err
}

func (s *MemoryStore) updateUser(id uint, u *User) error {
	// called with acquired mu lock
	if id != u.ID {
		return ErrUpdateID
	}
	if _, exists := s.users[u.ID]; !exists {
		return ErrNotFound
	}
	if eid, exists := s.emails[u.Email]; exists && eid != id {
		return ErrDup
	}
	s.users[u.ID] = *u
	return nil
}

func (s *MemoryStore) GetUser(id uint, u *User) error {
	s.mu.RLock()
	var exists bool
	*u, exists = s.users[id]
	if !exists {
		s.mu.RUnlock()
		return ErrNotFound
	}
	s.mu.RUnlock()
	return nil
}

func (s *MemoryStore) GetUserVisits(id uint, q *UserVisitsQuery, visits *[]UserVisit) error {
	s.mu.RLock()
	userVisits, ok := s.visitsByUser[id]
	if !ok {
		s.mu.RUnlock()
		return ErrNotFound
	}
	results := make([]UserVisit, 0, userVisits.Size())
	iterator := userVisits.Iterator()
	for iterator.Next() {
		visitedAt := iterator.Key().(int64)
		if (q.FromDate != nil && visitedAt <= *q.FromDate) ||
			(q.ToDate != nil && visitedAt >= *q.ToDate) {
			continue
		}
		visitID := iterator.Value().(uint)
		visit := s.visits[visitID]
		location := s.locations[visit.LocationID]
		if (q.Country != "" && location.Country != q.Country) ||
			(q.ToDistance != nil && *location.Distance >= *q.ToDistance) {
			continue
		}
		results = append(results, UserVisit{
			Mark:      *visit.Mark,
			VisitedAt: *visit.VisitedAt,
			Place:     location.Place,
		})
	}
	*visits = results
	s.mu.RUnlock()
	return nil
}

// Location methods
func (s *MemoryStore) CreateLocation(l *Location) error {
	s.mu.Lock()
	err := s.createLocation(l)
	s.mu.Unlock()
	return err
}

func (s *MemoryStore) CreateLocations(ls []Location) error {
	s.mu.Lock()
	var err error
	for _, l := range ls {
		err = s.createLocation(&l)
		if err != nil {
			break
		}
	}
	s.mu.Unlock()
	return err
}

func (s *MemoryStore) createLocation(l *Location) error {
	// called with acquired mu lock
	if l.ID == 0 {
		return ErrMissingID
	}
	if _, exists := s.locations[l.ID]; exists {
		return ErrDup
	}
	s.locations[l.ID] = *l
	s.visitsByLocation[l.ID] = redblacktree.NewWith(timestampComparator)
	return nil
}

func (s *MemoryStore) UpdateLocation(id uint, l *Location) error {
	s.mu.Lock()
	err := s.updateLocation(id, l)
	s.mu.Unlock()
	return err
}

func (s *MemoryStore) updateLocation(id uint, l *Location) error {
	// called with acquired mu lock
	if id != l.ID {
		return ErrUpdateID
	}
	if _, exists := s.locations[l.ID]; !exists {
		return ErrNotFound
	}
	s.locations[l.ID] = *l
	return nil
}

func (s *MemoryStore) GetLocation(id uint, l *Location) error {
	s.mu.RLock()
	var exists bool
	*l, exists = s.locations[id]
	if !exists {
		s.mu.RUnlock()
		return ErrNotFound
	}
	s.mu.RUnlock()
	return nil
}

func (s *MemoryStore) GetLocationAvg(id uint, q *LocationAvgQuery) (float64, error) {
	s.mu.RLock()
	locationVisits, ok := s.visitsByLocation[id]
	if !ok {
		s.mu.RUnlock()
		return 0, ErrNotFound
	}
	iterator := locationVisits.Iterator()
	var sum, cnt int
	for iterator.Next() {
		visitedAt := iterator.Key().(int64)
		if (q.FromDate != nil && visitedAt <= *q.FromDate) ||
			(q.ToDate != nil && visitedAt >= *q.ToDate) {
			continue
		}
		visitID := iterator.Value().(uint)
		visit := s.visits[visitID]
		if q.FromAge != nil || q.ToAge != nil || q.Gender != "" {
			user := s.users[visit.UserID]
			fromBirth := q.FromBirth()
			toBirth := q.ToBirth()
			if (fromBirth != nil && *user.BirthDate <= *fromBirth) ||
				(toBirth != nil && *user.BirthDate >= *toBirth) ||
				(q.Gender != "" && q.Gender != user.Gender) {
				continue
			}
		}

		sum += *visit.Mark
		cnt++
	}
	s.mu.RUnlock()

	var avg float64
	if cnt > 0 {
		avg = float64(sum) / float64(cnt)
	}
	return avg, nil
}

// Visit methods
func (s *MemoryStore) CreateVisit(v *Visit) error {
	s.mu.Lock()
	err := s.createVisit(v)
	s.mu.Unlock()
	return err
}

func (s *MemoryStore) CreateVisits(vs []Visit) error {
	s.mu.Lock()
	var err error
	for _, v := range vs {
		err = s.createVisit(&v)
		if err != nil {
			break
		}
	}
	s.mu.Unlock()
	return err
}

func (s *MemoryStore) createVisit(v *Visit) error {
	// called with acquired mu lock
	if v.ID == 0 {
		return ErrMissingID
	}
	if _, exists := s.visits[v.ID]; exists {
		return ErrDup
	}
	userVisits, ok := s.visitsByUser[v.UserID]
	if !ok {
		return ErrNotFound
	}
	locationVisits, ok := s.visitsByLocation[v.LocationID]
	if !ok {
		return ErrNotFound
	}

	s.visits[v.ID] = *v
	userVisits.Put(*v.VisitedAt, v.ID)
	locationVisits.Put(*v.VisitedAt, v.ID)
	return nil
}

func (s *MemoryStore) UpdateVisit(id uint, v *Visit) error {
	s.mu.Lock()
	err := s.updateVisit(id, v)
	s.mu.Unlock()
	return err
}

func (s *MemoryStore) updateVisit(id uint, v *Visit) error {
	// called with acquired mu lock
	if id != v.ID {
		return ErrUpdateID
	}
	prev, ok := s.visits[v.ID]
	if !ok {
		return ErrNotFound
	}
	s.visits[v.ID] = *v
	// update references
	if prev.UserID != v.UserID ||
		prev.VisitedAt != v.VisitedAt {
		// user index changed
		userVisits := s.visitsByUser[prev.UserID]
		userVisits.Remove(*prev.VisitedAt)
		if prev.UserID != v.UserID {
			userVisits = s.visitsByUser[v.UserID]
		}
		userVisits.Put(*v.VisitedAt, v.ID)
	}
	if prev.LocationID != v.LocationID ||
		prev.VisitedAt != v.VisitedAt {
		// location index changed
		locationVisits := s.visitsByLocation[prev.LocationID]
		locationVisits.Remove(*prev.VisitedAt)
		if prev.LocationID != v.LocationID {
			locationVisits = s.visitsByLocation[v.LocationID]
		}
		locationVisits.Put(*v.VisitedAt, v.ID)
	}
	return nil
}

func (s *MemoryStore) GetVisit(id uint, v *Visit) error {
	s.mu.RLock()
	var exists bool
	*v, exists = s.visits[id]
	if !exists {
		s.mu.RUnlock()
		return ErrNotFound
	}
	s.mu.RUnlock()
	return nil
}

func (s *MemoryStore) Clear() error {
	// Memory store is empty at start
	return nil
}

func timestampComparator(a, b interface{}) int {
	aTimestamp := a.(int64)
	bTimestamp := b.(int64)
	return int(aTimestamp - bTimestamp)
}
