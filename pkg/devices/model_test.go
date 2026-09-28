/*
 * Copyright (c) 2026 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package devices

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestFilterCriteriaGetAspectIds(t *testing.T) {
	cases := []struct {
		name     string
		criteria FilterCriteria
		expected []string
	}{
		{name: "returns an empty list when no aspect is set", criteria: FilterCriteria{FunctionId: "f"}, expected: []string{}},
		{name: "treats the deprecated aspect id as a single element list", criteria: FilterCriteria{AspectId: "a"}, expected: []string{"a"}},
		{name: "returns the aspect list unchanged", criteria: FilterCriteria{AspectIds: []string{"a", "b"}}, expected: []string{"a", "b"}},
		{name: "appends the deprecated aspect id to the list", criteria: FilterCriteria{AspectId: "c", AspectIds: []string{"a", "b"}}, expected: []string{"a", "b", "c"}},
		{name: "does not duplicate a deprecated aspect id already in the list", criteria: FilterCriteria{AspectId: "b", AspectIds: []string{"a", "b"}}, expected: []string{"a", "b"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := c.criteria.GetAspectIds()
			if !reflect.DeepEqual(actual, c.expected) {
				t.Errorf("expected %#v, got %#v", c.expected, actual)
			}
		})
	}
}

func TestFilterCriteriaGetAspectIdsDoesNotModifyTheCriteria(t *testing.T) {
	list := make([]string, 1, 2)
	list[0] = "a"
	criteria := FilterCriteria{AspectId: "b", AspectIds: list}
	criteria.GetAspectIds()
	if !reflect.DeepEqual(criteria, FilterCriteria{AspectId: "b", AspectIds: []string{"a"}}) {
		t.Errorf("criteria was modified: %#v", criteria)
	}
	if spare := list[:2]; spare[1] != "" {
		t.Errorf("backing array of AspectIds was written: %#v", spare)
	}
}

func TestFilterCriteriaDecodesAspectIds(t *testing.T) {
	var criteria []FilterCriteria
	err := json.Unmarshal([]byte(`[{"function_id":"f","aspect_ids":["a","b"]},{"function_id":"f","aspect_id":"c"}]`), &criteria)
	if err != nil {
		t.Fatal(err)
	}
	expected := []FilterCriteria{
		{FunctionId: "f", AspectIds: []string{"a", "b"}},
		{FunctionId: "f", AspectId: "c"},
	}
	if !reflect.DeepEqual(criteria, expected) {
		t.Errorf("expected %#v, got %#v", expected, criteria)
	}
}

func TestFilterCriteriaWithoutAspectListEncodesAsBefore(t *testing.T) {
	for _, list := range [][]string{nil, {}} {
		actual, err := json.Marshal(FilterCriteria{Interaction: EVENT, FunctionId: "f", AspectId: "a", AspectIds: list})
		if err != nil {
			t.Fatal(err)
		}
		expected := `{"interaction":"event","function_id":"f","device_class_id":"","aspect_id":"a"}`
		if string(actual) != expected {
			t.Errorf("expected %v, got %v", expected, string(actual))
		}
	}
}
