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

type Config struct {
	WorkerParamPrefix   string `json:"worker_param_prefix"`
	FlowEngineUrl       string `json:"flow_engine_url"`
	FlowParserUrl       string `json:"flow_parser_url"`
	ImportDeployUrl     string `json:"import_deploy_url"`
	DeviceRepositoryUrl string `json:"device_repository_url"`
	Debug               bool   `json:"debug"`

	EnableMultiplePaths bool   `json:"enable_multiple_paths"`
	DevicePathPrefix    string `json:"device_path_prefix"`
	GroupPathPrefix     string `json:"group_path_prefix"`
	ImportPathPrefix    string `json:"import_path_prefix"`

	RemoveImportPathRoot bool `json:"remove_import_path_root"`
}
