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
	"fmt"
	"github.com/SENERGY-Platform/smart-service-module-worker-analytics/pkg/auth"
	"github.com/SENERGY-Platform/smart-service-module-worker-analytics/pkg/configuration"
	"github.com/SENERGY-Platform/smart-service-module-worker-analytics/pkg/model"
	"io"
	"log"
	"net/http"
	"runtime/debug"
)

func New(config configuration.Config, auth *auth.Auth, smartServiceRepo SmartServiceRepo) *Analytics {
	return &Analytics{config: config, auth: auth, smartServiceRepo: smartServiceRepo}
}

type Analytics struct {
	config           configuration.Config
	auth             *auth.Auth
	smartServiceRepo SmartServiceRepo
}

type SmartServiceRepo interface {
	GetInstanceUser(instanceId string) (userId string, err error)
}

func (this *Analytics) Do(task model.CamundaExternalTask) (modules []model.Module, outputs map[string]interface{}, err error) {
	userId, err := this.smartServiceRepo.GetInstanceUser(task.ProcessInstanceId)
	if err != nil {
		log.Println("ERROR: unable to get instance user", err)
		return modules, outputs, err
	}
	token, err := this.auth.ExchangeUserToken(userId)
	if err != nil {
		log.Println("ERROR: unable to exchange user token", err)
		return modules, outputs, err
	}

	analyticsModuleData, analyticsModuleDeleteInfo, outputs, err := this.doAnalytics(token, task)
	if err != nil {
		return modules, outputs, err
	}
	moduleData := this.getModuleData(task)
	for key, value := range analyticsModuleData {
		moduleData[key] = value
	}

	return []model.Module{{
			Id:               this.getModuleId(task),
			ProcesInstanceId: task.ProcessInstanceId,
			SmartServiceModuleInit: model.SmartServiceModuleInit{
				DeleteInfo: analyticsModuleDeleteInfo,
				ModuleType: this.config.CamundaWorkerTopic,
				ModuleData: moduleData,
			},
		}},
		outputs,
		err
}

func (this *Analytics) Undo(modules []model.Module, reason error) {
	log.Println("UNDO:", reason)
	for _, module := range modules {
		if module.DeleteInfo != nil {
			err := this.useModuleDeleteInfo(*module.DeleteInfo)
			if err != nil {
				log.Println("ERROR:", err)
				debug.PrintStack()
			}
		}
	}
}

func (this *Analytics) useModuleDeleteInfo(info model.ModuleDeleteInfo) error {
	req, err := http.NewRequest("DELETE", info.Url, nil)
	if err != nil {
		return err
	}
	if info.UserId != "" {
		token, err := this.auth.ExchangeUserToken(info.UserId)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", token.Jwt())
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 && resp.StatusCode != http.StatusNotFound {
		temp, _ := io.ReadAll(resp.Body)
		err = fmt.Errorf("unexpected response: %v, %v", resp.StatusCode, string(temp))
		log.Println("ERROR:", err)
		debug.PrintStack()
		return err
	}
	_, _ = io.ReadAll(resp.Body)
	return nil
}

func (this *Analytics) getModuleId(task model.CamundaExternalTask) string {
	return task.ProcessInstanceId + "." + task.Id
}

func (this *Analytics) doAnalytics(token auth.Token, task model.CamundaExternalTask) (moduleData map[string]interface{}, deleteInfo *model.ModuleDeleteInfo, outputs map[string]interface{}, err error) {
	panic("not implemented")
	//TODO
}
