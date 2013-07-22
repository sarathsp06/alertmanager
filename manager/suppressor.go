// Copyright 2013 Prometheus Team
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package manager

import (
	"container/heap"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"
)

type Suppression struct {
	Id        uint
	CreatedBy string
	CreatedAt time.Time
	EndsAt    time.Time
	Comment   string
	Filters   Filters
}

type Suppressor struct {
	Suppressions Suppressions
	mu           sync.Mutex
}

type IsInhibitedInterrogator interface {
	IsInhibited(*Event) (bool, *Suppression)
}

func NewSuppressor() *Suppressor {
	return &Suppressor{}
}

type Suppressions []*Suppression

func (s Suppressions) Len() int {
	return len(s)
}

func (s Suppressions) Less(i, j int) bool {
	return s[i].EndsAt.Before(s[j].EndsAt)
}

func (s Suppressions) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s *Suppressions) Push(v interface{}) {
	*s = append(*s, v.(*Suppression))
}

func (s *Suppressions) Pop() interface{} {
	old := *s
	n := len(old)
	item := old[n-1]
	*s = old[0 : n-1]
	return item
}

func (s *Suppressor) nextSuppressionId() uint {
	// BUG: generate proper ID.
	return 1
}

func (s *Suppressor) AddSuppression(sup *Suppression) uint {
	s.mu.Lock()
	defer s.mu.Unlock()

	sup.Id = s.nextSuppressionId()
	heap.Push(&s.Suppressions, sup)
	return sup.Id
}

func (s *Suppressor) UpdateSuppression(sup *Suppression) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldSup, present := s.suppressionById(sup.Id)
	if !present {
		return fmt.Errorf("Suppression with ID %d doesn't exist", sup.Id)
	}
	*oldSup = *sup
	return nil
}

func (s *Suppressor) GetSuppression(id uint) (*Suppression, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sup, present := s.suppressionById(id)
	if !present {
		return nil, fmt.Errorf("Suppression with ID %d doesn't exist", id)
	}
	return sup, nil
}

func (s *Suppressor) DelSuppression(id uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, sup := range s.Suppressions {
		if sup.Id == id {
			s.Suppressions = append(s.Suppressions[:i], s.Suppressions[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("Suppression with ID %d doesn't exist", id)
}

func (s *Suppressor) SuppressionSummary() Suppressions {
	s.mu.Lock()
	defer s.mu.Unlock()

	suppressions := make(Suppressions, 0, len(s.Suppressions))
	for _, sup := range s.Suppressions {
		suppressions = append(suppressions, sup)
	}
	return suppressions
}

func (s *Suppressor) IsInhibited(e *Event) (bool, *Suppression) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, s := range s.Suppressions {
		if s.Filters.Handles(e) {
			return true, s
		}
	}
	return false, nil
}

func (s *Suppressor) suppressionById(id uint) (*Suppression, bool) {
	// BUG: use a separate index for ID lookups once this becomes necessary.
	for _, sup := range s.Suppressions {
		if sup.Id == id {
			return sup, true
		}
	}
	return nil, false
}

func (s *Suppressor) reapSuppressions(t time.Time) {
	log.Println("reaping suppression...")

	i := sort.Search(len(s.Suppressions), func(i int) bool {
		return (s.Suppressions)[i].EndsAt.After(t)
	})

	s.Suppressions = s.Suppressions[i:]

	// BUG(matt): Validate if strictly necessary.
	heap.Init(&s.Suppressions)
}
