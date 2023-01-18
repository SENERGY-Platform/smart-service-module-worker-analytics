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
	"github.com/SENERGY-Platform/smart-service-module-worker-analytics/pkg/devices"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/auth"
	"github.com/SENERGY-Platform/smart-service-module-worker-lib/pkg/model"
	"log"
)

func (this *Analytics) getServicesAndPathsForGroupSelection(token auth.Token, selection model.DeviceGroupSelection, criteria []devices.FilterCriteria) (serviceIds []string, serviceToDevices map[string][]string, serviceToPath map[string][]string, err error) {
	devices, deviceTypeIds, err := this.devices.GetDeviceInfosOfGroup(token, selection.Id)
	if err != nil {
		return nil, nil, nil, err
	}
	return this.getServicesAndPathsForDevices(token, devices, deviceTypeIds, criteria)
}

func (this *Analytics) getServicesAndPathsForDeviceIdList(token auth.Token, deviceIds []string, criteria []devices.FilterCriteria) (serviceIds []string, serviceToDevices map[string][]string, serviceToPath map[string][]string, err error) {
	devices, deviceTypeIds, err := this.devices.GetDeviceInfosOfDevices(token, deviceIds)
	if err != nil {
		return nil, nil, nil, err
	}
	return this.getServicesAndPathsForDevices(token, devices, deviceTypeIds, criteria)
}

func (this *Analytics) getServicesAndPathsForDevices(token auth.Token, deviceList []devices.Device, deviceTypeIds []string, criteria []devices.FilterCriteria) (serviceIds []string, serviceToDevices map[string][]string, serviceToPath map[string][]string, err error) {
	options, err := this.getDeviceGroupPathOptions(token, criteria, deviceTypeIds)
	if err != nil {
		log.Println("ERROR: unable to find path options", err)
		return nil, nil, nil, err
	}
	serviceIds = []string{}
	serviceToDevices = map[string][]string{}
	serviceToPath = map[string][]string{}
	serviceToPathToCharacteristic := map[string]map[string]string{}
	for _, device := range deviceList {
		for _, option := range options[device.DeviceTypeId] {
			if len(option.JsonPath) > 0 {
				serviceToDevices[option.ServiceId] = append(serviceToDevices[option.ServiceId], device.Id)
				if _, ok := serviceToPath[option.ServiceId]; !ok {
					serviceIds = append(serviceIds, option.ServiceId)
				}
				for _, path := range option.JsonPath {
					serviceToPath[option.ServiceId] = append(serviceToPath[option.ServiceId], path)
					if _, ok := serviceToPathToCharacteristic[option.ServiceId]; !ok {
						serviceToPathToCharacteristic[option.ServiceId] = map[string]string{}
					}
					serviceToPathToCharacteristic[option.ServiceId][path] = option.PathToCharacteristicId[path]
				}
			}
		}
	}
	return serviceIds, serviceToDevices, serviceToPath, nil
}

func (this *Analytics) getDeviceGroupPathOptions(token auth.Token, criteria []devices.FilterCriteria, deviceTypeIds []string) (result map[string][]devices.PathOptionsResultElement, err error) {
	result = map[string][]devices.PathOptionsResultElement{}
	for i, c := range criteria {
		if c.Interaction == "" {
			c.Interaction = devices.EVENT
		}
		criteria[i] = c
	}
	selectables, err := this.devices.GetDeviceTypeSelectables(token, criteria, true, true)
	if err != nil {
		return result, err
	}
	for _, dtId := range deviceTypeIds {
		for _, selectable := range selectables {
			if selectable.DeviceTypeId == dtId {
				for sid, options := range selectable.ServicePathOptions {
					temp := devices.PathOptionsResultElement{
						ServiceId:              sid,
						JsonPath:               []string{},
						PathToCharacteristicId: map[string]string{},
					}
					for _, option := range options {
						if option.ServiceId == sid {
							temp.JsonPath = append(temp.JsonPath, option.Path)
							temp.PathToCharacteristicId[option.Path] = option.CharacteristicId
						} else {
							log.Println("WARNING: unexpected service id in ServicePathOptions")
						}
					}
					result[dtId] = append(result[dtId], temp)
				}
			}
		}
	}
	return result, nil
}
