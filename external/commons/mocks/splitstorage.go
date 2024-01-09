package mocks

import (
	"github.com/splitio/go-split-commons/v5/dtos"
	"github.com/splitio/go-split-commons/v5/storage"
	"github.com/splitio/go-toolkit/v5/datastructures/set"
	"github.com/stretchr/testify/mock"
)

type SplitStorageMock struct{ mock.Mock }

func (m *SplitStorageMock) All() []dtos.SplitDTO {
	args := m.Called()
	return args.Get(0).([]dtos.SplitDTO)
}

func (m *SplitStorageMock) ChangeNumber() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *SplitStorageMock) FetchMany(names []string) map[string]*dtos.SplitDTO {
	args := m.Called(names)
	return args.Get(0).(map[string]*dtos.SplitDTO)
}

func (m *SplitStorageMock) KillLocally(name string, defaultTreatment string, newCn int64) {
	m.Called(name, defaultTreatment, newCn)
}

func (m *SplitStorageMock) SegmentNames() *set.ThreadUnsafeSet {
	args := m.Called()
	return args.Get(0).(*set.ThreadUnsafeSet)
}

func (m *SplitStorageMock) SetChangeNumber(changeNumber int64) error {
	args := m.Called(changeNumber)
	return args.Error(0)

}

func (m *SplitStorageMock) Split(splitName string) *dtos.SplitDTO {
	args := m.Called(splitName)
	return args.Get(0).(*dtos.SplitDTO)
}

func (m *SplitStorageMock) SplitNames() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *SplitStorageMock) Update(toAdd []dtos.SplitDTO, toRemove []dtos.SplitDTO, newCN int64) {
	m.Called()
}

func (m *SplitStorageMock) TrafficTypeExists(trafficType string) bool {
	args := m.Called(trafficType)
	return args.Bool(0)
}

func (m *SplitStorageMock) GetNamesByFlagSets(sets []string) map[string][]string {
	args := m.Called(sets)
	return args.Get(0).(map[string][]string)
}

var _ storage.SplitStorage = (*SplitStorageMock)(nil)
