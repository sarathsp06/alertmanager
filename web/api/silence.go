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

package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"code.google.com/p/gorest"

	"github.com/prometheus/alert_manager/manager"
)

type Filter struct {
	NamePattern string
	ValuePattern string
}

type Silence struct {
	CreatedBy string
	CreatedAt int64
	EndsAt    int64
	Comment   string
	Filters   []*Filter
}

func (s AlertManagerService) AddSilence(sc Silence) {
	/*
	endsAt, err := strconv.Atoi(sc.EndsAt)
	if err != nil {
		rb := s.ResponseBuilder()
		rb.SetResponseCode(http.StatusBadRequest)
		//return err.Error()
		return
	}
	createdAt, err := strconv.Atoi(sc.CreatedAt)
	if err != nil {
		rb := s.ResponseBuilder()
		rb.SetResponseCode(http.StatusBadRequest)
		//return err.Error()
		return
	}*/
	filters := make(manager.Filters, 0, len(sc.Filters))
	for _, f := range sc.Filters {
		filters = append(filters, manager.NewFilter(f.NamePattern, f.ValuePattern))
	}

	sup := &manager.Suppression{
		CreatedBy: sc.CreatedBy,
		CreatedAt: time.Unix(sc.CreatedAt, 0),
		EndsAt: time.Unix(sc.EndsAt, 0),
		Comment: sc.Comment,
		Filters: filters,
	}
	s.Suppressor.AddSuppression(sup)
}

func (s AlertManagerService) SilenceSummary() string {
	rb := s.ResponseBuilder()
	rb.SetContentType(gorest.Application_Json)
	silenceSummary := s.Suppressor.SuppressionSummary()
	resultBytes, err := json.Marshal(silenceSummary)
	if err != nil {
		log.Printf("Error marshalling metric names: %v", err)
		rb.SetResponseCode(http.StatusInternalServerError)
		return err.Error()
	}
	return string(resultBytes)
}
