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
	"github.com/SENERGY-Platform/smart-service-module-worker-analytics/pkg/model"
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
