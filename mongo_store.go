package main

import (
	"errors"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var ErrUpdateID = errors.New("id field cannot be changed")

type sessionFunc func(s *mgo.Session) error

type MongoStore struct {
	s *mgo.Session
}

func NewMongoStore(s *mgo.Session) *MongoStore {
	return &MongoStore{s}
}

// User methods
func (s *MongoStore) CreateUser(u *User) error {
	return s.withSession(func(s *mgo.Session) error {
		return usersCollection(s).Insert(u)
	})
}

func (s *MongoStore) UpdateUser(id int, u *User) error {
	if id != u.ID {
		return ErrUpdateID
	}
	return s.withSession(func(s *mgo.Session) error {
		return usersCollection(s).UpdateId(id, u)
	})
}

func (s *MongoStore) GetUser(id int, u *User) error {
	return s.withSession(func(s *mgo.Session) error {
		return usersCollection(s).FindId(id).One(u)
	})
}

func (s *MongoStore) GetUserVisits(id int, visits *[]Visit) error {
	return s.withSession(func(s *mgo.Session) error {
		return visitsCollection(s).Find(bson.M{"u": id}).Sort("v").All(visits)
	})
}

// Location methods
func (s *MongoStore) CreateLocation(l *Location) error {
	return s.withSession(func(s *mgo.Session) error {
		return locationsCollection(s).Insert(l)
	})
}

func (s *MongoStore) UpdateLocation(id int, l *Location) error {
	if id != l.ID {
		return ErrUpdateID
	}
	return s.withSession(func(s *mgo.Session) error {
		return locationsCollection(s).UpdateId(id, l)
	})
}

func (s *MongoStore) GetLocation(id int, l *Location) error {
	return s.withSession(func(s *mgo.Session) error {
		return locationsCollection(s).FindId(id).One(l)
	})
}

func (s *MongoStore) GetLocationAvg(id int) (float32, error) {
	// TODO: aggregation
	return 0, nil
}

// Visit methods
func (s *MongoStore) CreateVisit(v *Visit) error {
	return s.withSession(func(s *mgo.Session) error {
		return visitsCollection(s).Insert(v)
	})
}

func (s *MongoStore) UpdateVisit(id int, v *Visit) error {
	if id != v.ID {
		return ErrUpdateID
	}
	return s.withSession(func(s *mgo.Session) error {
		return visitsCollection(s).UpdateId(id, v)
	})
}

func (s *MongoStore) GetVisit(id int, v *Visit) error {
	return s.withSession(func(s *mgo.Session) error {
		return visitsCollection(s).FindId(id).One(v)
	})
}

func (s *MongoStore) Clear() error {
	return s.withSession(func(s *mgo.Session) error {
		if _, err := usersCollection(s).RemoveAll(nil); err != nil {
			return err
		}
		if _, err := locationsCollection(s).RemoveAll(nil); err != nil {
			return err
		}
		if _, err := visitsCollection(s).RemoveAll(nil); err != nil {
			return err
		}
		return nil
	})
}

func (s *MongoStore) withSession(f sessionFunc) error {
	session := s.s.Clone() // wrap session
	err := f(session)
	session.Close()
	return err
}

func usersCollection(s *mgo.Session) *mgo.Collection {
	return s.DB("").C("users")
}

func locationsCollection(s *mgo.Session) *mgo.Collection {
	return s.DB("").C("locations")
}

func visitsCollection(s *mgo.Session) *mgo.Collection {
	return s.DB("").C("visits")
}
