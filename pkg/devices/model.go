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

package devices

type FilterCriteria struct {
	FunctionId    string `json:"function_id"`
	DeviceClassId string `json:"device_class_id"`
	AspectId      string `json:"aspect_id"`
}

type Device struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	DeviceTypeId string `json:"device_type_id"`
}

type DeviceGroup struct {
	Id        string   `json:"id"`
	Name      string   `json:"name"`
	DeviceIds []string `json:"device_ids"`
}

type PathOptionsResultElement struct {
	ServiceId              string            `json:"service_id"`
	JsonPath               []string          `json:"json_path"`
	PathToCharacteristicId map[string]string `json:"path_to_characteristic_id"`
}

type DeviceTypeSelectable struct {
	DeviceTypeId string `json:"device_type_id,omitempty"`
	//Services           []Service                      `json:"services,omitempty"`
	ServicePathOptions map[string][]ServicePathOption `json:"service_path_options,omitempty"`
}

type ServicePathOption struct {
	ServiceId        string `json:"service_id"`
	Path             string `json:"path"`
	CharacteristicId string `json:"characteristic_id"`
	//AspectNode            AspectNode     `json:"aspect_node"`
	FunctionId            string      `json:"function_id"`
	IsVoid                bool        `json:"is_void"`
	Value                 interface{} `json:"value,omitempty"`
	IsControllingFunction bool        `json:"is_controlling_function"`
	//Configurables         []Configurable `json:"configurables,omitempty"`
	//Type                  Type           `json:"type,omitempty"`
}
