package main

import "github.com/stretchr/testify/mock"

type MockStore struct {
	mock.Mock
}

func (m *MockStore) CreateUser(u *User) error {
	return m.Called(u).Error(0)
}

func (m *MockStore) CreateUsers(us []User) error {
	return m.Called(us).Error(0)
}

func (m *MockStore) UpdateUser(id uint, u *User) error {
	return m.Called(id, u).Error(0)
}

func (m *MockStore) GetUser(id uint, u *User) error {
	return m.Called(id, u).Error(0)
}

func (m *MockStore) GetUserVisits(id uint, q *UserVisitsQuery, visits *[]UserVisit) error {
	return m.Called(id, q, visits).Error(0)
}

func (m *MockStore) CreateLocation(l *Location) error {
	return m.Called(l).Error(0)
}

func (m *MockStore) CreateLocations(ls []Location) error {
	return m.Called(ls).Error(0)
}

func (m *MockStore) UpdateLocation(id uint, l *Location) error {
	return m.Called(id, l).Error(0)
}

func (m *MockStore) GetLocation(id uint, l *Location) error {
	return m.Called(id, l).Error(0)
}

func (m *MockStore) GetLocationAvg(id uint, q *LocationAvgQuery) (float64, error) {
	args := m.Called(id, q)
	avg, _ := args.Get(0).(float64)
	return avg, args.Error(1)
}

func (m *MockStore) CreateVisit(v *Visit) error {
	return m.Called(v).Error(0)
}

func (m *MockStore) CreateVisits(vs []Visit) error {
	return m.Called(vs).Error(0)
}

func (m *MockStore) UpdateVisit(id uint, v *Visit) error {
	return m.Called(id, v).Error(0)
}

func (m *MockStore) GetVisit(id uint, v *Visit) error {
	return m.Called(id, v).Error(0)
}

func (m *MockStore) Clear() error {
	return m.Called().Error(0)
}
