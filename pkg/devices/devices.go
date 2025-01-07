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

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/auth"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
)

type Devices struct {
	deviceRepositoryUrl string
}

func New(deviceRepositoryUrl string) *Devices {
	return &Devices{deviceRepositoryUrl: deviceRepositoryUrl}
}

func (this *Devices) GetDeviceInfosOfGroup(token auth.Token, groupId string) (devices []Device, deviceTypeIds []string, err error) {
	group, err := this.GetDeviceGroup(token, groupId)
	if err != nil {
		return devices, nil, err
	}
	return this.GetDeviceInfosOfDevices(token, group.DeviceIds)
}

func (this *Devices) GetDeviceInfosOfDevices(token auth.Token, deviceIds []string) (devices []Device, deviceTypeIds []string, err error) {
	devices, err = this.GetDevicesWithIds(token, deviceIds)
	if err != nil {
		return devices, nil, err
	}
	deviceTypeIsUsed := map[string]bool{}
	for _, d := range devices {
		if !deviceTypeIsUsed[d.DeviceTypeId] {
			deviceTypeIsUsed[d.DeviceTypeId] = true
			deviceTypeIds = append(deviceTypeIds, d.DeviceTypeId)
		}
	}
	return devices, deviceTypeIds, nil
}

func (this *Devices) GetDeviceGroup(token auth.Token, groupId string) (result DeviceGroup, err error) {
	dg, err, _ := client.NewClient(this.deviceRepositoryUrl, nil).ReadDeviceGroup(groupId, token.Jwt(), false)
	if err != nil {
		return result, err
	}
	return DeviceGroup{
		Id:        dg.Id,
		Name:      dg.Name,
		DeviceIds: dg.DeviceIds,
	}, nil
}

func (this *Devices) GetDevicesWithIds(token auth.Token, ids []string) (result []Device, err error) {
	device, err, _ := client.NewClient(this.deviceRepositoryUrl, nil).ListDevices(token.Jwt(), client.DeviceListOptions{
		Ids:    ids,
		Limit:  int64(len(ids)),
		Offset: 0,
		SortBy: "name.asc",
	})
	if err != nil {
		return result, err
	}
	for _, d := range device {
		result = append(result, Device{
			Id:           d.Id,
			Name:         d.Name,
			DeviceTypeId: d.DeviceTypeId,
		})
	}
	return result, nil
}

func (this *Devices) GetDeviceTypeSelectables(token auth.Token, criteria []FilterCriteria, includeModified bool, servicesMustMatchAllCriteria bool) (result []DeviceTypeSelectable, err error) {
	requestBody := new(bytes.Buffer)
	err = json.NewEncoder(requestBody).Encode(criteria)
	if err != nil {
		return result, err
	}
	query := url.Values{}
	query.Set("services_must_match_all_criteria", strconv.FormatBool(servicesMustMatchAllCriteria))
	query.Set("include_id_modified", strconv.FormatBool(includeModified))
	req, err := http.NewRequest("POST", this.deviceRepositoryUrl+"/v2/query/device-type-selectables?"+query.Encode(), requestBody)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	req.Header.Set("Authorization", token.Jwt())
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		err = errors.New(buf.String())
		log.Println("ERROR: ", resp.StatusCode, err)
		debug.PrintStack()
		return result, err
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		debug.PrintStack()
		return result, err
	}

	return result, nil
}
