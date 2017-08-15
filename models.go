package main

import (
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

	t.Time = time.Unix(int64(ts), 0)

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

func (t *Timestamp) String() string {
	return time.Time(t.Time).String()
}

func ValidateTimestamp(field reflect.Value) interface{} {
	if timestamp, ok := field.Interface().(Timestamp); ok {
		if !timestamp.Time.IsZero() {
			return timestamp.Time.Unix()
		}
	}
	return nil
}
