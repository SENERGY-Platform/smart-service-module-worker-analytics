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
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"

	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/auth"
)

var DefaultTimeout = 30 * time.Second

func (this *Analytics) SendDeployRequest(token auth.Token, request PipelineRequest) (result Pipeline, err error, code int) {
	body, err := json.Marshal(request)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	this.libConfig.GetLogger().Debug("deploy event pipeline", "request", string(body))
	client := http.Client{
		Timeout: DefaultTimeout,
	}
	req, err := http.NewRequest(
		"POST",
		this.config.FlowEngineUrl+"/pipeline",
		bytes.NewBuffer(body),
	)
	if err != nil {
		this.libConfig.GetLogger().Error("error in SendDeployRequest", "error", err, "stack", string(debug.Stack()))
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token.Jwt())
	req.Header.Set("X-UserId", token.GetUserId())
	this.libConfig.GetLogger().Debug("send analytics deployment with token", "token", req.Header.Get("Authorization"))
	resp, err := client.Do(req)
	if err != nil {
		this.libConfig.GetLogger().Error("error in SendDeployRequest", "error", err, "stack", string(debug.Stack()))
		return result, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		err = errors.New("unexpected statuscode")
		this.libConfig.GetLogger().Error("error in SendDeployRequest", "error", err, "stack", string(debug.Stack()), "statuscode", resp.StatusCode)
		return result, err, resp.StatusCode
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err, http.StatusOK
}

func (this *Analytics) SendUpdateRequest(token auth.Token, request PipelineRequest) (result Pipeline, err error, code int) {
	body, err := json.Marshal(request)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	this.libConfig.GetLogger().Debug("deploy event pipeline", "request", string(body))
	client := http.Client{
		Timeout: DefaultTimeout,
	}
	req, err := http.NewRequest(
		"PUT",
		this.config.FlowEngineUrl+"/pipeline",
		bytes.NewBuffer(body),
	)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token.Jwt())
	req.Header.Set("X-UserId", token.GetUserId())
	this.libConfig.GetLogger().Debug("send analytics deployment update with token", "token", req.Header.Get("Authorization"))
	resp, err := client.Do(req)
	if err != nil {
		this.libConfig.GetLogger().Error("error in SendDeployRequest", "error", err, "stack", string(debug.Stack()))
		return result, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		err = errors.New("unexpected statuscode")
		this.libConfig.GetLogger().Error("error in SendDeployRequest", "error", err, "stack", string(debug.Stack()), "statuscode", resp.StatusCode)
		return result, err, resp.StatusCode
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err, http.StatusOK
}

func (this *Analytics) Remove(token auth.Token, pipelineId string) error {
	client := http.Client{
		Timeout: DefaultTimeout,
	}
	req, err := http.NewRequest(
		"DELETE",
		this.config.FlowEngineUrl+"/pipeline/"+url.PathEscape(pipelineId),
		nil,
	)
	if err != nil {
		this.libConfig.GetLogger().Error("error in Remove", "error", err, "stack", string(debug.Stack()))
		return err
	}
	req.Header.Set("Authorization", token.Jwt())
	req.Header.Set("X-UserId", token.GetUserId())
	resp, err := client.Do(req)
	if err != nil {
		this.libConfig.GetLogger().Error("error in Remove", "error", err, "stack", string(debug.Stack()))
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		err = errors.New("unexpected statuscode")
		this.libConfig.GetLogger().Error("error in Remove", "error", err, "stack", string(debug.Stack()), "statuscode", resp.StatusCode)
		return err
	}
	return nil
}

func (this *Analytics) GetFlowInputs(token auth.Token, id string) (result []FlowModelCell, err error, code int) {
	client := http.Client{
		Timeout: DefaultTimeout,
	}
	req, err := http.NewRequest(
		"GET",
		this.config.FlowParserUrl+"/flow/getinputs/"+url.PathEscape(id),
		nil,
	)
	if err != nil {
		this.libConfig.GetLogger().Error("error in GetFlowInputs", "error", err, "stack", string(debug.Stack()))
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token.Jwt())
	req.Header.Set("X-UserId", token.GetUserId())
	resp, err := client.Do(req)
	if err != nil {
		this.libConfig.GetLogger().Error("error in GetFlowInputs", "error", err, "stack", string(debug.Stack()))
		return result, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		err = errors.New("unexpected statuscode")
		this.libConfig.GetLogger().Error("error in GetFlowInputs", "error", err, "stack", string(debug.Stack()), "statuscode", resp.StatusCode)
		return result, err, resp.StatusCode
	}

	temp, err := io.ReadAll(resp.Body)
	err = json.Unmarshal(temp, &result)
	if err != nil {
		this.libConfig.GetLogger().Error("error in GetFlowInputs", "error", err, "stack", string(debug.Stack()), "payload", string(temp))
		return result, err, http.StatusInternalServerError
	}
	return result, err, http.StatusOK
}
