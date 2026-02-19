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

package pkg

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SENERGY-Platform/smart-service-module-worker-analytics/pkg/analytics"
	"github.com/SENERGY-Platform/smart-service-module-worker-analytics/pkg/devices"
	"github.com/SENERGY-Platform/smart-service-module-worker-analytics/pkg/imports"
	lib "github.com/SENERGY-Platform/smart-service-module-worker-lib"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/auth"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/camunda"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/configuration"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/model"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/smartservicerepository"
)

func getPipelineId(moduleData map[string]interface{}) (string, error) {
	pipelineId, ok := moduleData["pipeline_id"].(string)
	if ok {
		return pipelineId, nil
	}
	pipeline, ok := moduleData["pipeline"]
	if !ok {
		return "", fmt.Errorf("missing pipeline_id or pipeline in module data")
	}
	pipelineObj, ok := pipeline.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid pipeline in module data")
	}
	pipelineId, ok = pipelineObj["id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid pipeline in module data (id is not string)")
	}
	return pipelineId, nil
}

func Start(ctx context.Context, wg *sync.WaitGroup, config analytics.Config, libConfig configuration.Config) error {
	handlerFactory := func(auth *auth.Auth, smartServiceRepo *smartservicerepository.SmartServiceRepository) (camunda.Handler, error) {
		handler := analytics.New(
			config,
			libConfig,
			auth,
			smartServiceRepo,
			imports.New(config.ImportDeployUrl),
			devices.New(config.DeviceRepositoryUrl),
		)
		interval, err := time.ParseDuration(config.HealthCheckInterval)
		if err != nil {
			return nil, err
		}

		healthCheck := func(module model.SmartServiceModule) (health error, err error) {
			token, err := auth.ExchangeUserToken(module.UserId)
			if err != nil {
				return nil, err
			}
			pipelineId, err := getPipelineId(module.ModuleData)
			if err != nil {
				return nil, err
			}
			state, code, err := handler.CheckPipeline(token, pipelineId)
			if err != nil {
				if code == 0 {
					return nil, err
				}
				return err, nil
			}
			if !state.Running {
				return fmt.Errorf("pipeline not running"), nil
			}
			return nil, nil
		}
		moduleQuery := model.ModulQuery{TypeFilter: &libConfig.CamundaWorkerTopic}
		smartServiceRepo.StartHealthCheck(ctx, interval, moduleQuery, healthCheck) //timer loop
		smartServiceRepo.RunHealthCheck(moduleQuery, healthCheck)                  //initial check
		return handler, nil
	}
	return lib.Start(ctx, wg, libConfig, handlerFactory)
}
