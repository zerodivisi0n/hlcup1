package main

import "github.com/mailru/easyjson"

type JSONProxy struct {
	store Store
}

func NewJSONProxy(store Store) *JSONProxy {
	return &JSONProxy{store}
}

// User methods
func (p JSONProxy) CreateUser(u *User) error {
	data, err := easyjson.Marshal(u)
	if err != nil {
		return err
	}
	u.JSON = data
	return p.store.CreateUser(u)
}

func (p JSONProxy) CreateUsers(us []User) error {
	for i, u := range us {
		data, err := easyjson.Marshal(u)
		if err != nil {
			return err
		}
		us[i].JSON = data
	}
	return p.store.CreateUsers(us)
}

func (p JSONProxy) UpdateUser(id uint, u *User) error {
	data, err := easyjson.Marshal(u)
	if err != nil {
		return err
	}
	u.JSON = data
	return p.store.UpdateUser(id, u)
}

func (p JSONProxy) GetUser(id uint, u *User) error {
	return p.store.GetUser(id, u)
}

func (p JSONProxy) GetUserVisits(id uint, q *UserVisitsQuery, visits *[]UserVisit) error {
	return p.store.GetUserVisits(id, q, visits)
}

// Location methods
func (p JSONProxy) CreateLocation(l *Location) error {
	data, err := easyjson.Marshal(l)
	if err != nil {
		return err
	}
	l.JSON = data
	return p.store.CreateLocation(l)
}

func (p JSONProxy) CreateLocations(ls []Location) error {
	for i, l := range ls {
		data, err := easyjson.Marshal(l)
		if err != nil {
			return err
		}
		ls[i].JSON = data
	}
	return p.store.CreateLocations(ls)
}

func (p JSONProxy) UpdateLocation(id uint, l *Location) error {
	data, err := easyjson.Marshal(l)
	if err != nil {
		return err
	}
	l.JSON = data
	return p.store.UpdateLocation(id, l)
}

func (p JSONProxy) GetLocation(id uint, l *Location) error {
	return p.store.GetLocation(id, l)
}

func (p JSONProxy) GetLocationAvg(id uint, q *LocationAvgQuery) (float64, error) {
	return p.store.GetLocationAvg(id, q)
}

// Visit methods
func (p JSONProxy) CreateVisit(v *Visit) error {
	data, err := easyjson.Marshal(v)
	if err != nil {
		return err
	}
	v.JSON = data
	return p.store.CreateVisit(v)
}

func (p JSONProxy) CreateVisits(vs []Visit) error {
	for i, v := range vs {
		data, err := easyjson.Marshal(v)
		if err != nil {
			return err
		}
		vs[i].JSON = data
	}
	return p.store.CreateVisits(vs)
}

func (p JSONProxy) UpdateVisit(id uint, v *Visit) error {
	data, err := easyjson.Marshal(v)
	if err != nil {
		return err
	}
	v.JSON = data
	return p.store.UpdateVisit(id, v)
}

func (p JSONProxy) GetVisit(id uint, v *Visit) error {
	return p.store.GetVisit(id, v)
}

func (p JSONProxy) Clear() error {
	return p.store.Clear()
}
