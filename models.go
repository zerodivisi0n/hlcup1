package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type User struct {
	ID        int       `json:"id" bson:"_id" validate:"omitempty,gt=0"`
	FirstName string    `json:"first_name" bson:"f" validate:"omitempty,max=50"`
	LastName  string    `json:"last_name" bson:"l" validate:"omitempty,max=50"`
	Email     string    `json:"email" bson:"e" validate:"omitempty,email,max=100"`
	Gender    string    `json:"gender" bson:"g" validate:"omitempty,eq=m|eq=f"`
	BirthDate Timestamp `json:"birth_date" bson:"b" validate:"omitempty,gte=-1262304000,lte=915148800"` // 01.01.1930 - 01.01.1999
}

type Location struct {
	ID       int    `json:"id" bson:"_id" validate:"omitempty,gt=0"`
	City     string `json:"city" bson:"ci" validate:"omitempty,max=50"`
	Country  string `json:"country" bson:"co" validate:"omitempty,max=50"`
	Place    string `json:"place" bson:"p"`
	Distance int    `json:"distance" bson:"d" validate:"omitempty,gte=0"`
}

type Visit struct {
	ID         int       `json:"id" bson:"_id" validate:"omitempty,gt=0"`
	UserID     int       `json:"user" bson:"u" validate:"omitempty,gt=0"`
	LocationID int       `json:"location" bson:"l" validate:"omitempty,gt=0"`
	VisitedAt  Timestamp `json:"visited_at" bson:"v" validate:"omitempty,gte=946684800,lte=1420070400"` // 01.01.2000 - 01.01.2015
	Mark       int       `json:"mark" bson:"m" validate:"omitempty,gte=0,lte=5"`
}

type UserVisit struct {
	Mark      int       `json:"mark" bson:"m"`
	VisitedAt Timestamp `json:"visited_at" bson:"v"`
	Place     string    `json:"place" bson:"p"`
}

type UserVisitsQuery struct {
	FromDate   Timestamp `schema:"fromDate"`
	ToDate     Timestamp `schema:"toDate"`
	Country    string    `schema:"country"`
	ToDistance int       `schema:"toDistance"`
}

type LocationAvgQuery struct {
	FromDate Timestamp `schema:"fromDate"`
	ToDate   Timestamp `schema:"toDate"`
	FromAge  int       `schema:"fromAge"`
	ToAge    int       `schema:"toAge"`
	Gender   string    `schema:"gender" validate:"omitempty,eq=m|eq=f"`
}

type Timestamp struct {
	time.Time
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	ts := t.Time.Unix()
	stamp := fmt.Sprint(ts)

	return []byte(stamp), nil
}

func (t *Timestamp) UnmarshalJSON(b []byte) error {
	ts, err := strconv.Atoi(string(b))
	if err != nil {
		return err
	}

	t.SetUnix(int64(ts))

	return nil
}

func (t Timestamp) GetBSON() (interface{}, error) {
	if time.Time(t.Time).IsZero() {
		return nil, nil
	}

	return time.Time(t.Time), nil
}

func (t *Timestamp) SetBSON(raw bson.Raw) error {
	var tm time.Time

	if err := raw.Unmarshal(&tm); err != nil {
		return err
	}

	*t = Timestamp{tm}

	return nil
}

func (t *Timestamp) UnmarshalText(text []byte) error {
	return t.UnmarshalJSON(text)
}

func (t *Timestamp) String() string {
	return time.Time(t.Time).String()
}

func (t *Timestamp) SetUnix(ts int64) {
	t.Time = time.Unix(ts, 0)
}

func ValidateTimestamp(field reflect.Value) interface{} {
	if timestamp, ok := field.Interface().(Timestamp); ok {
		if !timestamp.Time.IsZero() {
			return timestamp.Time.Unix()
		}
	}
	return nil
}

// Custom unmarshalers
func (u *User) UnmarshalJSON(b []byte) error {
	kv := map[string]interface{}{}
	if err := json.Unmarshal(b, &kv); err != nil {
		return err
	}
	for k, v := range kv {
		var ok bool
		switch k {
		case "id":
			var id float64
			if id, ok = v.(float64); ok {
				u.ID = int(id)
			}
		case "first_name":
			u.FirstName, ok = v.(string)
		case "last_name":
			u.LastName, ok = v.(string)
		case "email":
			u.Email, ok = v.(string)
		case "gender":
			u.Gender, ok = v.(string)
		case "birth_date":
			var ts float64
			if ts, ok = v.(float64); ok {
				u.BirthDate.SetUnix(int64(ts))
			}
		}
		if !ok {
			return fmt.Errorf("Invalid type %T for key '%s'", v, k)
		}
	}
	return nil
}

func (l *Location) UnmarshalJSON(b []byte) error {
	kv := map[string]interface{}{}
	if err := json.Unmarshal(b, &kv); err != nil {
		return err
	}
	for k, v := range kv {
		var ok bool
		switch k {
		case "id":
			var id float64
			if id, ok = v.(float64); ok {
				l.ID = int(id)
			}
		case "city":
			l.City, ok = v.(string)
		case "country":
			l.Country, ok = v.(string)
		case "place":
			l.Place, ok = v.(string)
		case "distance":
			var d float64
			if d, ok = v.(float64); ok {
				l.Distance = int(d)
			}
		}
		if !ok {
			return fmt.Errorf("Invalid type %T for key '%s'", v, k)
		}
	}
	return nil
}

func (vi *Visit) UnmarshalJSON(b []byte) error {
	kv := map[string]interface{}{}
	if err := json.Unmarshal(b, &kv); err != nil {
		return err
	}
	for k, v := range kv {
		var ok bool
		switch k {
		case "id":
			var id float64
			if id, ok = v.(float64); ok {
				vi.ID = int(id)
			}
		case "user":
			var id float64
			if id, ok = v.(float64); ok {
				vi.UserID = int(id)
			}
		case "location":
			var id float64
			if id, ok = v.(float64); ok {
				vi.LocationID = int(id)
			}
		case "visited_at":
			var ts float64
			if ts, ok = v.(float64); ok {
				vi.VisitedAt.SetUnix(int64(ts))
			}
		case "mark":
			var mark float64
			if mark, ok = v.(float64); ok {
				vi.Mark = int(mark)
			}
		}
		if !ok {
			return fmt.Errorf("Invalid type %T for key '%s'", v, k)
		}
	}
	return nil
}
