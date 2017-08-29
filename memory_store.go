package main

import (
	"sync"

	"github.com/emirpasic/gods/trees/redblacktree"
)

type MemoryStore struct {
	mu               sync.RWMutex
	users            []*User
	locations        []*Location
	visits           []*Visit
	emails           map[string]uint
	visitsByUser     []*redblacktree.Tree
	visitsByLocation []*redblacktree.Tree
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:            make([]*User, 10000),
		locations:        make([]*Location, 10000),
		visits:           make([]*Visit, 10000),
		emails:           make(map[string]uint, 10000),
		visitsByUser:     make([]*redblacktree.Tree, 10000),
		visitsByLocation: make([]*redblacktree.Tree, 10000),
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
	curLen := len(s.users)
	intID := int(u.ID)
	if curLen <= intID {
		newLen := curLen
		if intID > newLen {
			newLen = intID
		}
		s.users = append(s.users, make([]*User, newLen+1000)...)
		s.visitsByUser = append(s.visitsByUser, make([]*redblacktree.Tree, newLen+1000)...)
	}
	if s.users[u.ID] != nil {
		return ErrDup
	}
	if _, exists := s.emails[u.Email]; exists {
		return ErrDup
	}
	uCopy := *u
	s.users[u.ID] = &uCopy
	s.emails[u.Email] = u.ID
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
	if len(s.users) <= int(id) || s.users[id] == nil {
		return ErrNotFound
	}
	if eid, exists := s.emails[u.Email]; exists && eid != id {
		return ErrDup
	}
	prev := s.users[id]
	if prev.Email != u.Email {
		delete(s.emails, prev.Email)
		s.emails[u.Email] = u.ID
	}
	*s.users[u.ID] = *u
	return nil
}

func (s *MemoryStore) GetUser(id uint, u *User) error {
	s.mu.RLock()
	if len(s.users) <= int(id) || s.users[id] == nil {
		s.mu.RUnlock()
		return ErrNotFound
	}
	*u = *s.users[id]
	s.mu.RUnlock()
	return nil
}

func (s *MemoryStore) GetUserVisits(id uint, q *UserVisitsQuery, visits *[]UserVisit) error {
	s.mu.RLock()
	if len(s.visitsByUser) <= int(id) || s.visitsByUser[id] == nil {
		s.mu.RUnlock()
		return ErrNotFound
	}
	userVisits := s.visitsByUser[id]
	results := make([]UserVisit, 0, userVisits.Size())
	iterator := userVisits.Iterator()
	for iterator.Next() {
		visitedAt := iterator.Key().(int64)
		if (q.FromDate != nil && visitedAt <= *q.FromDate) ||
			(q.ToDate != nil && visitedAt >= *q.ToDate) {
			continue
		}
		visit := iterator.Value().(*Visit)
		location := s.locations[visit.LocationID]
		if (q.Country != "" && location.Country != q.Country) ||
			(q.ToDistance != nil && location.Distance >= *q.ToDistance) {
			continue
		}
		results = append(results, UserVisit{
			Mark:      visit.Mark,
			VisitedAt: visit.VisitedAt,
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
	curLen := len(s.locations)
	intID := int(l.ID)
	if curLen <= intID {
		newLen := curLen
		if intID > newLen {
			newLen = intID
		}
		s.locations = append(s.locations, make([]*Location, newLen+1000)...)
		s.visitsByLocation = append(s.visitsByLocation, make([]*redblacktree.Tree, newLen+1000)...)
	}
	if s.locations[l.ID] != nil {
		return ErrDup
	}
	lCopy := *l
	s.locations[l.ID] = &lCopy
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
	if len(s.locations) <= int(id) || s.locations[id] == nil {
		return ErrNotFound
	}
	*s.locations[l.ID] = *l
	return nil
}

func (s *MemoryStore) GetLocation(id uint, l *Location) error {
	s.mu.RLock()
	if len(s.locations) <= int(id) || s.locations[id] == nil {
		s.mu.RUnlock()
		return ErrNotFound
	}
	*l = *s.locations[id]
	s.mu.RUnlock()
	return nil
}

func (s *MemoryStore) GetLocationAvg(id uint, q *LocationAvgQuery) (float64, error) {
	s.mu.RLock()
	if len(s.visitsByLocation) <= int(id) || s.visitsByLocation[id] == nil {
		s.mu.RUnlock()
		return 0, ErrNotFound
	}
	locationVisits := s.visitsByLocation[id]
	iterator := locationVisits.Iterator()
	var sum, cnt int
	for iterator.Next() {
		visitedAt := iterator.Key().(int64)
		if (q.FromDate != nil && visitedAt <= *q.FromDate) ||
			(q.ToDate != nil && visitedAt >= *q.ToDate) {
			continue
		}
		visit := iterator.Value().(*Visit)
		if q.FromAge != nil || q.ToAge != nil || q.Gender != "" {
			user := s.users[visit.UserID]
			fromBirth := q.FromBirth()
			toBirth := q.ToBirth()
			if (fromBirth != nil && user.BirthDate <= *fromBirth) ||
				(toBirth != nil && user.BirthDate >= *toBirth) ||
				(q.Gender != "" && q.Gender != user.Gender) {
				continue
			}
		}

		sum += visit.Mark
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
	}
	s.mu.Unlock()
	return err
}

func (s *MemoryStore) createVisit(v *Visit) error {
	// called with acquired mu lock
	if v.ID == 0 {
		return ErrMissingID
	}
	curLen := len(s.visits)
	intID := int(v.ID)
	if curLen <= intID {
		newLen := curLen
		if intID > newLen {
			newLen = intID
		}
		s.visits = append(s.visits, make([]*Visit, newLen+1000)...)
	}
	if s.visits[v.ID] != nil {
		return ErrDup
	}
	if len(s.visitsByUser) <= int(v.UserID) || s.visitsByUser[v.UserID] == nil {
		return ErrNotFound
	}
	if len(s.visitsByLocation) <= int(v.LocationID) || s.visitsByLocation[v.LocationID] == nil {
		return ErrNotFound
	}
	vCopy := *v
	s.visits[v.ID] = &vCopy
	s.visitsByUser[v.UserID].Put(v.VisitedAt, &vCopy)
	s.visitsByLocation[v.LocationID].Put(v.VisitedAt, &vCopy)
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
	if len(s.visits) <= int(id) || s.visits[id] == nil {
		return ErrNotFound
	}
	// update references
	cur := s.visits[v.ID]
	if cur.UserID != v.UserID ||
		cur.VisitedAt != v.VisitedAt {
		// user index changed
		userVisits := s.visitsByUser[cur.UserID]
		userVisits.Remove(cur.VisitedAt)
		if cur.UserID != v.UserID {
			userVisits = s.visitsByUser[v.UserID]
		}
		userVisits.Put(v.VisitedAt, cur)
	}
	if cur.LocationID != v.LocationID ||
		cur.VisitedAt != v.VisitedAt {
		// location index changed
		locationVisits := s.visitsByLocation[cur.LocationID]
		locationVisits.Remove(cur.VisitedAt)
		if cur.LocationID != v.LocationID {
			locationVisits = s.visitsByLocation[v.LocationID]
		}
		locationVisits.Put(v.VisitedAt, cur)
	}
	*s.visits[v.ID] = *v
	return nil
}

func (s *MemoryStore) GetVisit(id uint, v *Visit) error {
	s.mu.RLock()
	if len(s.visits) <= int(id) || s.visits[id] == nil {
		s.mu.RUnlock()
		return ErrNotFound
	}
	*v = *s.visits[id]
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
