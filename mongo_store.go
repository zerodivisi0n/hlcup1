package main

import (
	"errors"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	ErrMissingID = errors.New("missing id")
	ErrUpdateID  = errors.New("id field cannot be changed")
)

type sessionFunc func(s *mgo.Session) error

type MongoStore struct {
	s *mgo.Session
}

func NewMongoStore(s *mgo.Session) *MongoStore {
	return &MongoStore{s}
}

// User methods
func (s *MongoStore) CreateUser(u *User) error {
	if u.ID == 0 {
		return ErrMissingID
	}
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

func (s *MongoStore) GetUserVisits(id int, q *UserVisitsQuery, visits *[]UserVisit) error {
	return s.withSession(func(s *mgo.Session) error {
		// Check users exists
		c, err := usersCollection(s).FindId(id).Count()
		if err != nil {
			return err
		}
		if c == 0 {
			return mgo.ErrNotFound
		}
		// Query visits
		return visitsCollection(s).Pipe(userVisitsPipeline(id, q)).All(visits)
	})
}

// Location methods
func (s *MongoStore) CreateLocation(l *Location) error {
	if l.ID == 0 {
		return ErrMissingID
	}
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

func (s *MongoStore) GetLocationAvg(id int, q *LocationAvgQuery) (float64, error) {
	var avg float64
	if err := s.withSession(func(s *mgo.Session) error {
		// Check location exists
		c, err := locationsCollection(s).FindId(id).Count()
		if err != nil {
			return nil
		}
		if c == 0 {
			return mgo.ErrNotFound
		}
		result := bson.M{}
		if err := visitsCollection(s).Pipe(locationAvgPipeline(id, q)).One(&result); err != nil {
			return nil
		}
		avg = result["avg"].(float64)
		return nil
	}); err != nil {
		return 0, nil
	}
	return avg, nil
}

// Visit methods
func (s *MongoStore) CreateVisit(v *Visit) error {
	if v.ID == 0 {
		return ErrMissingID
	}
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

func userVisitsPipeline(id int, q *UserVisitsQuery) []bson.M {
	matchStage := bson.M{"u": id}
	if tr := timeRangeQuery(q.FromDate.Time, q.ToDate.Time); tr != nil {
		matchStage["v"] = tr
	}

	filterStage := bson.M{}
	if q.Country != "" {
		filterStage["loc.co"] = q.Country
	}
	if q.ToDistance != 0 {
		filterStage["loc.d"] = bson.M{"$lt": q.ToDistance}
	}

	return []bson.M{
		{"$match": matchStage},                                                                          // filter by user
		{"$sort": bson.M{"v": 1}},                                                                       // ascending order
		{"$lookup": bson.M{"from": "locations", "localField": "l", "foreignField": "_id", "as": "loc"}}, // join location
		{"$unwind": "$loc"},                                           // unwind location array
		{"$match": filterStage},                                       // filter results by country and distance
		{"$project": bson.M{"_id": 0, "m": 1, "v": 1, "p": "$loc.p"}}, // build result
	}
}

func locationAvgPipeline(id int, q *LocationAvgQuery) []bson.M {
	matchStage := bson.M{"l": id}
	if tr := timeRangeQuery(q.FromDate.Time, q.ToDate.Time); tr != nil {
		matchStage["v"] = tr
	}

	groupStage := bson.M{"_id": "_", "avg": bson.M{"$avg": "$m"}}
	if q.FromAge == 0 && q.ToAge == 0 && q.Gender == "" {
		return []bson.M{
			{"$match": matchStage},
			{"$group": groupStage},
		}
	}

	// add lookup stage
	filterStage := bson.M{}
	var fromBirth, toBirth time.Time
	if q.FromAge > 0 {
		toBirth = time.Now().AddDate(-q.FromAge, 0, 0)
	}
	if q.ToAge > 0 {
		fromBirth = time.Now().AddDate(-q.ToAge, 0, 0)
	}

	if tr := timeRangeQuery(fromBirth, toBirth); tr != nil {
		filterStage["user.b"] = tr
	}
	if q.Gender != "" {
		filterStage["user.g"] = q.Gender
	}

	return []bson.M{
		{"$match": matchStage},
		{"$lookup": bson.M{"from": "users", "localField": "u", "foreignField": "_id", "as": "user"}},
		{"$unwind": "$user"},
		{"$match": filterStage},
		{"$group": groupStage},
	}
}

func timeRangeQuery(from, to time.Time) bson.M {
	if !from.IsZero() && !to.IsZero() {
		return bson.M{
			"$gt": from,
			"$lt": to,
		}
	} else if !from.IsZero() {
		return bson.M{"$gt": from}
	} else if !to.IsZero() {
		return bson.M{"$lt": to}
	}
	return nil
}
