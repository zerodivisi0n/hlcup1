package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type sessionFunc func(s *mgo.Session) error

type MongoStore struct {
	s *mgo.Session
}

func NewMongoStore(s *mgo.Session) (*MongoStore, error) {
	if err := usersCollection(s).EnsureIndex(mgo.Index{
		Key:    []string{"e"},
		Unique: true,
	}); err != nil {
		return nil, err
	}
	if err := visitsCollection(s).EnsureIndexKey("u", "v"); err != nil {
		return nil, err
	}
	if err := visitsCollection(s).EnsureIndexKey("l", "v"); err != nil {
		return nil, err
	}
	return &MongoStore{s}, nil
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

func (s *MongoStore) CreateUsers(us []User) error {
	docs := make([]interface{}, len(us))
	for i, u := range us {
		if u.ID == 0 {
			return ErrMissingID
		}
		docs[i] = u
	}
	return s.withSession(func(s *mgo.Session) error {
		bulk := usersCollection(s).Bulk()
		bulk.Insert(docs...)
		_, err := bulk.Run()
		return err
	})
}

func (s *MongoStore) UpdateUser(id uint, u *User) error {
	if id != u.ID {
		return ErrUpdateID
	}
	return s.withSession(func(s *mgo.Session) error {
		return usersCollection(s).UpdateId(id, u)
	})
}

func (s *MongoStore) GetUser(id uint, u *User) error {
	return s.withSession(func(s *mgo.Session) error {
		return usersCollection(s).FindId(id).One(u)
	})
}

func (s *MongoStore) GetUserVisits(id uint, q *UserVisitsQuery, visits *[]UserVisit) error {
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

func (s *MongoStore) CreateLocations(ls []Location) error {
	docs := make([]interface{}, len(ls))
	for i, l := range ls {
		if l.ID == 0 {
			return ErrMissingID
		}
		docs[i] = l
	}
	return s.withSession(func(s *mgo.Session) error {
		bulk := locationsCollection(s).Bulk()
		bulk.Insert(docs...)
		_, err := bulk.Run()
		return err
	})
}

func (s *MongoStore) UpdateLocation(id uint, l *Location) error {
	if id != l.ID {
		return ErrUpdateID
	}
	return s.withSession(func(s *mgo.Session) error {
		return locationsCollection(s).UpdateId(id, l)
	})
}

func (s *MongoStore) GetLocation(id uint, l *Location) error {
	return s.withSession(func(s *mgo.Session) error {
		return locationsCollection(s).FindId(id).One(l)
	})
}

func (s *MongoStore) GetLocationAvg(id uint, q *LocationAvgQuery) (float64, error) {
	var avg float64
	if err := s.withSession(func(s *mgo.Session) error {
		// Check location exists
		c, err := locationsCollection(s).FindId(id).Count()
		if err != nil {
			return err
		}
		if c == 0 {
			return mgo.ErrNotFound
		}
		result := bson.M{}
		err = visitsCollection(s).Pipe(locationAvgPipeline(id, q)).One(&result)
		if err == mgo.ErrNotFound {
			return nil
		} else if err != nil {
			return err
		}
		avg = result["avg"].(float64)
		return nil
	}); err != nil {
		return 0, err
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

func (s *MongoStore) CreateVisits(vs []Visit) error {
	docs := make([]interface{}, len(vs))
	for i, v := range vs {
		if v.ID == 0 {
			return ErrMissingID
		}
		docs[i] = v
	}
	return s.withSession(func(s *mgo.Session) error {
		bulk := visitsCollection(s).Bulk()
		bulk.Insert(docs...)
		_, err := bulk.Run()
		return err
	})
}

func (s *MongoStore) UpdateVisit(id uint, v *Visit) error {
	if id != v.ID {
		return ErrUpdateID
	}
	return s.withSession(func(s *mgo.Session) error {
		return visitsCollection(s).UpdateId(id, v)
	})
}

func (s *MongoStore) GetVisit(id uint, v *Visit) error {
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
	if mgo.IsDup(err) {
		err = ErrDup
	} else if err == mgo.ErrNotFound {
		return ErrNotFound
	}
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

func userVisitsPipeline(id uint, q *UserVisitsQuery) []bson.M {
	matchStage := bson.M{"u": id}
	if tr := timeRangeQuery(q.FromDate, q.ToDate); tr != nil {
		matchStage["v"] = tr
	}

	filterStage := bson.M{}
	if q.Country != "" {
		filterStage["loc.co"] = q.Country
	}
	if q.ToDistance != nil {
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

func locationAvgPipeline(id uint, q *LocationAvgQuery) []bson.M {
	matchStage := bson.M{"l": id}
	if tr := timeRangeQuery(q.FromDate, q.ToDate); tr != nil {
		matchStage["v"] = tr
	}

	groupStage := bson.M{"_id": "_", "avg": bson.M{"$avg": "$m"}}
	if q.FromAge == nil && q.ToAge == nil && q.Gender == "" {
		return []bson.M{
			{"$match": matchStage},
			{"$group": groupStage},
		}
	}

	// add lookup stage
	filterStage := bson.M{}
	if tr := timeRangeQuery(q.FromBirth(), q.ToBirth()); tr != nil {
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

func timeRangeQuery(from, to *int64) bson.M {
	if from != nil && to != nil {
		return bson.M{
			"$gt": int64(*from),
			"$lt": int64(*to),
		}
	} else if from != nil {
		return bson.M{"$gt": *from}
	} else if to != nil {
		return bson.M{"$lt": *to}
	}
	return nil
}
