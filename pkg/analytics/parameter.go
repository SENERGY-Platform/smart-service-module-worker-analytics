/*
 * Copyright (c) 2022 InfAI (CC SES)
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

package analytics

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/smart-service-module-worker-analytics/pkg/devices"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/model"
	"strconv"
	"strings"
)

func (this *Analytics) getModuleData(task model.CamundaExternalTask) (result map[string]interface{}) {
	result = map[string]interface{}{}
	variable, ok := task.Variables[this.config.WorkerParamPrefix+"module_data"]
	if !ok {
		return result
	}
	str, ok := variable.Value.(string)
	if !ok {
		return result
	}
	err := json.Unmarshal([]byte(str), &result)
	if err != nil {
		return map[string]interface{}{}
	}
	return result
}

func (this *Analytics) getPipelineName(task model.CamundaExternalTask) string {
	variable, ok := task.Variables[this.config.WorkerParamPrefix+"name"]
	if !ok {
		return ""
	}
	result, ok := variable.Value.(string)
	if !ok {
		return ""
	}
	return result
}

func (this *Analytics) getConsumeAllMessages(task model.CamundaExternalTask) (result bool) {
	variable, ok := task.Variables[this.config.WorkerParamPrefix+"consume_all_messages"]
	if !ok {
		return false
	}
	switch v := variable.Value.(type) {
	case string:
		err := json.Unmarshal([]byte(v), &result)
		if err != nil {
			return false
		}
		return result
	case bool:
		return v
	default:
		return false
	}
}

func (this *Analytics) getPipelineWindowTime(task model.CamundaExternalTask) (int, error) {
	variable, ok := task.Variables[this.config.WorkerParamPrefix+"window_time"]
	if !ok {
		return 0, nil
	}
	switch value := variable.Value.(type) {
	case string:
		temp, err := strconv.Atoi(value)
		if err != nil {
			return 0, err
		}
		return temp, nil
	case float64:
		return int(value), nil
	case int:
		return int(value), nil
	case int64:
		return int(value), nil
	default:
		return 0, errors.New("unknown type for window_time")
	}
}

func (this *Analytics) getPipelineMergeStrategy(task model.CamundaExternalTask) (string, error) {
	variable, ok := task.Variables[this.config.WorkerParamPrefix+"merge_strategy"]
	if !ok {
		return "inner", nil
	}
	value, ok := variable.Value.(string)
	if !ok {
		return "inner", errors.New("expected string for merge_strategy")
	}
	if value == "" {
		return "inner", nil
	}
	return value, nil
}

func (this *Analytics) getPipelineDescription(task model.CamundaExternalTask) string {
	variable, ok := task.Variables[this.config.WorkerParamPrefix+"desc"]
	if !ok {
		return ""
	}
	result, ok := variable.Value.(string)
	if !ok {
		return ""
	}
	return result
}

func (this *Analytics) getFlowId(task model.CamundaExternalTask) string {
	variable, ok := task.Variables[this.config.WorkerParamPrefix+"flow_id"]
	if !ok {
		return ""
	}
	result, ok := variable.Value.(string)
	if !ok {
		return ""
	}
	return result
}

func (this *Analytics) getPersistData(task model.CamundaExternalTask, inputId string) (result bool) {
	variable, ok := task.Variables[this.config.WorkerParamPrefix+"persistData."+inputId]
	if !ok {
		return false
	}
	str, ok := variable.Value.(string)
	if !ok {
		return false
	}
	err := json.Unmarshal([]byte(str), &result)
	if err != nil {
		return false
	}
	return result
}

func (this *Analytics) getPipelineNodeConfig(task model.CamundaExternalTask, inputId string, confName string) (string, error) {
	key := this.config.WorkerParamPrefix + "conf." + inputId + "." + confName
	variable, ok := task.Variables[key]
	if !ok {
		return "", nil //errors.New("missing pipeline input config (" + this.config.WorkerParamPrefix + "conf." + inputId + "." + confName + ")")
	}
	if strings.ToLower(variable.Type) == "null" {
		return "", nil
	}
	if variable.Value == nil {
		return "", nil
	}
	switch v := variable.Value.(type) {
	case string:
		return v, nil
	case map[string]interface{}:
		temp, err := json.Marshal(v["value"])
		if err != nil {
			return "", fmt.Errorf("unable to interpret pipeline input config %v (%v) \n%w", key, v, err)
		}
		return string(temp), err
	default:
		temp, err := json.Marshal(variable.Value)
		if err != nil {
			return "", fmt.Errorf("unable to interpret pipeline input config %v (%v) \n%w", key, v, err)
		}
		return string(temp), err
	}
}

func (this *Analytics) getSelection(task model.CamundaExternalTask, inputId string, portName string) (result model.IotOption, err error) {
	variableName := this.config.WorkerParamPrefix + "selection." + inputId + "." + portName
	variable, ok := task.Variables[variableName]
	if !ok {
		return result, errors.New("missing pipeline input selection (" + variableName + ")")
	}
	selectionStr, ok := variable.Value.(string)
	if !ok {
		return result, errors.New("unable to interpret pipeline input selection (" + variableName + ")")
	}
	err = json.Unmarshal([]byte(selectionStr), &result)
	if err != nil {
		return result, fmt.Errorf("unable to interpret pipeline input selection (%v): %w", variableName, err)
	}
	return result, nil
}

func (this *Analytics) getNodePathCriteria(task model.CamundaExternalTask, inputId string, portName string) (result []devices.FilterCriteria, err error) {
	variableName := this.config.WorkerParamPrefix + "criteria." + inputId + "." + portName
	variable, ok := task.Variables[variableName]
	if !ok {
		return result, errors.New("missing pipeline input criteria (mandatory when selection is group or device without service) (" + variableName + ")")
	}
	criteriaStr, ok := variable.Value.(string)
	if !ok {
		return result, errors.New("unable to interpret pipeline input criteria (" + variableName + ")")
	}
	err = json.Unmarshal([]byte(criteriaStr), &result)
	if err != nil {
		return result, fmt.Errorf("unable to interpret pipeline input criteria (%v): %w", variableName, err)
	}
	return result, err
}

func (this *Analytics) getNodeServiceCriteria(task model.CamundaExternalTask, inputId string, portName string) (result []devices.FilterCriteria, err error) {
	variableName := this.config.WorkerParamPrefix + "service_criteria." + inputId + "." + portName
	variable, ok := task.Variables[variableName]
	if !ok {
		return nil, nil
	}
	criteriaStr, ok := variable.Value.(string)
	if !ok {
		return result, errors.New("unable to interpret pipeline input criteria (" + variableName + ")")
	}
	if criteriaStr == "" {
		return nil, nil
	}
	err = json.Unmarshal([]byte(criteriaStr), &result)
	if err != nil {
		return result, fmt.Errorf("unable to interpret pipeline input criteria (%v): %w", variableName, err)
	}
	return result, err
}

// if no key is set: return nil
func (this *Analytics) getModuleKey(task model.CamundaExternalTask) (key *string) {
	variable, ok := task.Variables[this.config.WorkerParamPrefix+"key"]
	if !ok {
		return nil
	}
	result, ok := variable.Value.(string)
	if ok {
		return &result
	}
	return nil
}
