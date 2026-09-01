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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"

	"github.com/SENERGY-Platform/gin-middleware/otelx"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/auth"
)

var DefaultTimeout = 30 * time.Second

func (this *Analytics) SendDeployRequest(ctx context.Context, token auth.Token, request PipelineRequest) (result Pipeline, err error, code int) {
	body, err := json.Marshal(request)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	this.libConfig.GetLogger().DebugContext(ctx, "deploy event pipeline", "request", string(body))
	client := http.Client{
		Timeout: DefaultTimeout,
	}
	req, err := http.NewRequest(
		"POST",
		this.config.FlowEngineUrl+"/pipeline",
		bytes.NewBuffer(body),
	)
	if err != nil {
		this.libConfig.GetLogger().ErrorContext(ctx, "error in SendDeployRequest", "error", err, "stack", string(debug.Stack()))
		return result, err, http.StatusInternalServerError
	}
	err = otelx.InjectContextToRequest(ctx, req)
	if err != nil {
		this.libConfig.GetLogger().ErrorContext(ctx, "error in SendDeployRequest", "error", err, "stack", string(debug.Stack()))
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token.Jwt())
	req.Header.Set("X-UserId", token.GetUserId())
	this.libConfig.GetLogger().DebugContext(ctx, "send analytics deployment")
	resp, err := client.Do(req)
	if err != nil {
		this.libConfig.GetLogger().ErrorContext(ctx, "error in SendDeployRequest", "error", err, "stack", string(debug.Stack()))
		return result, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		err = errors.New("unexpected statuscode")
		this.libConfig.GetLogger().ErrorContext(ctx, "error in SendDeployRequest", "error", err, "stack", string(debug.Stack()), "statuscode", resp.StatusCode)
		return result, err, resp.StatusCode
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err, http.StatusOK
}

func (this *Analytics) SendUpdateRequest(ctx context.Context, token auth.Token, request PipelineRequest) (result Pipeline, err error, code int) {
	body, err := json.Marshal(request)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	this.libConfig.GetLogger().DebugContext(ctx, "deploy event pipeline", "request", string(body))
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
	err = otelx.InjectContextToRequest(ctx, req)
	if err != nil {
		this.libConfig.GetLogger().ErrorContext(ctx, "error in SendUpdateRequest", "error", err, "stack", string(debug.Stack()))
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token.Jwt())
	req.Header.Set("X-UserId", token.GetUserId())
	this.libConfig.GetLogger().DebugContext(ctx, "send analytics deployment update")
	resp, err := client.Do(req)
	if err != nil {
		this.libConfig.GetLogger().ErrorContext(ctx, "error in SendDeployRequest", "error", err, "stack", string(debug.Stack()))
		return result, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		err = errors.New("unexpected statuscode")
		this.libConfig.GetLogger().ErrorContext(ctx, "error in SendDeployRequest", "error", err, "stack", string(debug.Stack()), "statuscode", resp.StatusCode)
		return result, err, resp.StatusCode
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err, http.StatusOK
}

func (this *Analytics) Remove(ctx context.Context, token auth.Token, pipelineId string) error {
	client := http.Client{
		Timeout: DefaultTimeout,
	}
	req, err := http.NewRequest(
		"DELETE",
		this.config.FlowEngineUrl+"/pipeline/"+url.PathEscape(pipelineId),
		nil,
	)
	if err != nil {
		this.libConfig.GetLogger().ErrorContext(ctx, "error in Remove", "error", err, "stack", string(debug.Stack()))
		return err
	}
	err = otelx.InjectContextToRequest(ctx, req)
	if err != nil {
		this.libConfig.GetLogger().ErrorContext(ctx, "error in Remove", "error", err, "stack", string(debug.Stack()))
		return err
	}
	req.Header.Set("Authorization", token.Jwt())
	req.Header.Set("X-UserId", token.GetUserId())
	resp, err := client.Do(req)
	if err != nil {
		this.libConfig.GetLogger().ErrorContext(ctx, "error in Remove", "error", err, "stack", string(debug.Stack()))
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		err = errors.New("unexpected statuscode")
		this.libConfig.GetLogger().ErrorContext(ctx, "error in Remove", "error", err, "stack", string(debug.Stack()), "statuscode", resp.StatusCode)
		return err
	}
	return nil
}

type PipelineState struct {
	Message       string `json:"message"`
	Name          string `json:"name"`
	Running       bool   `json:"running"`
	Transitioning bool   `json:"transitioning"`
}

func (this *Analytics) CheckPipeline(ctx context.Context, token auth.Token, pipelineId string) (state PipelineState, code int, err error) {
	client := http.Client{
		Timeout: DefaultTimeout,
	}
	req, err := http.NewRequest(
		"GET",
		this.config.FlowEngineUrl+"/pipeline/"+url.PathEscape(pipelineId),
		nil,
	)
	if err != nil {
		this.libConfig.GetLogger().ErrorContext(ctx, "error in CheckPipeline", "error", err, "stack", string(debug.Stack()))
		return state, 0, err
	}
	err = otelx.InjectContextToRequest(ctx, req)
	if err != nil {
		this.libConfig.GetLogger().ErrorContext(ctx, "error in CheckPipeline", "error", err, "stack", string(debug.Stack()))
		return state, 0, err
	}
	req.Header.Set("Authorization", token.Jwt())
	req.Header.Set("X-UserId", token.GetUserId())

	this.libConfig.GetLogger().DebugContext(ctx, "check pipeline request", "url", req.URL.String(), "method", req.Method, "xuser", req.Header.Get("X-UserId"))

	resp, err := client.Do(req)
	if err != nil {
		this.libConfig.GetLogger().ErrorContext(ctx, "error in CheckPipeline", "error", err, "stack", string(debug.Stack()))
		return state, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		pl, _ := io.ReadAll(resp.Body)
		err = fmt.Errorf("unexpected statuscode while checking pipeline %v: %v, %v", pipelineId, resp.StatusCode, string(pl))
		return state, resp.StatusCode, err
	}
	err = json.NewDecoder(resp.Body).Decode(&state)
	if err != nil {
		return state, 0, err
	}
	return state, resp.StatusCode, nil
}

func (this *Analytics) GetFlowInputs(ctx context.Context, token auth.Token, id string) (result []FlowModelCell, err error, code int) {
	client := http.Client{
		Timeout: DefaultTimeout,
	}
	req, err := http.NewRequest(
		"GET",
		this.config.FlowParserUrl+"/flow/getinputs/"+url.PathEscape(id),
		nil,
	)
	if err != nil {
		this.libConfig.GetLogger().ErrorContext(ctx, "error in GetFlowInputs", "error", err, "stack", string(debug.Stack()))
		return result, err, http.StatusInternalServerError
	}
	err = otelx.InjectContextToRequest(ctx, req)
	if err != nil {
		this.libConfig.GetLogger().ErrorContext(ctx, "error in GetFlowInputs", "error", err, "stack", string(debug.Stack()))
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token.Jwt())
	req.Header.Set("X-UserId", token.GetUserId())
	resp, err := client.Do(req)
	if err != nil {
		this.libConfig.GetLogger().ErrorContext(ctx, "error in GetFlowInputs", "error", err, "stack", string(debug.Stack()))
		return result, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		err = errors.New("unexpected statuscode")
		this.libConfig.GetLogger().ErrorContext(ctx, "error in GetFlowInputs", "error", err, "stack", string(debug.Stack()), "statuscode", resp.StatusCode)
		return result, err, resp.StatusCode
	}

	temp, err := io.ReadAll(resp.Body)
	err = json.Unmarshal(temp, &result)
	if err != nil {
		this.libConfig.GetLogger().ErrorContext(ctx, "error in GetFlowInputs", "error", err, "stack", string(debug.Stack()), "payload", string(temp))
		return result, err, http.StatusInternalServerError
	}
	return result, err, http.StatusOK
}
