/*
 * Copyright 2021 InfAI (CC SES)
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

package imports

import (
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/auth"
	"io"
	"net/http"
	"net/url"
)

type Imports struct {
	importDeployUrl string
}

func New(importDeployUrl string) *Imports {
	return &Imports{importDeployUrl: importDeployUrl}
}

func (this *Imports) GetTopic(token auth.Token, importId string) (topic string, err error) {
	req, err := http.NewRequest("GET", this.importDeployUrl+"/instances/"+url.PathEscape(importId), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", token.Jwt())
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		temp, _ := io.ReadAll(resp.Body)
		err = errors.New(string(temp))
		return "", err
	}
	var importInstance Import
	err = json.NewDecoder(resp.Body).Decode(&importInstance)
	topic = importInstance.KafkaTopic
	return topic, err
}

type Import struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	ImportTypeId string `json:"import_type_id"`
	Image        string `json:"image"`
	KafkaTopic   string `json:"kafka_topic"`
	//Configs      []ImportConfig `json:"configs"`
	Restart *bool `json:"restart"`
}
