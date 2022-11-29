## Outputs

### Pipeline-Id

- Desc: id of created pipeline
- Variable-Name: pipeline_id

## Camunda-Input-Variables

### Key
- Desc: identifies module for (later) update
- Variable-Name-Template: `{{config.WorkerParamPrefix}}.key`
- Variable-Name-Example: `analytics.key`
- Value: string

### Flow-Id

- Desc: defines which flow should be de deployed
- Variable-Name-Template: `{{config.WorkerParamPrefix}}.flow_id`
- Variable-Name-Example: `analytics.flow_id`
- Value: string

### Module-Data

- Desc: sets fields for Module.ModuleData
- Variable-Name-Template: `{{config.WorkerParamPrefix}}.module_data`
- Variable-Name-Example: `analytics.module_data`
- Value: `json.Marshal(map[string]interface{})`

### Analytics-Name

- Desc: sets the name of the analytics flow/pipeline deployment
- Variable-Name-Template: `{{config.WorkerParamPrefix}}.name`
- Variable-Name-Example: `analytics.name`
- Value: string

### Analytics-Description

- Desc: sets the description of the analytics flow/pipeline deployment
- Variable-Name-Template: `{{config.WorkerParamPrefix}}.desc`
- Variable-Name-Example: `analytics.desc`
- Value: string


### Window-Time

- Desc: sets the window-time of the analytics flow/pipeline deployment
- Variable-Name-Template: `{{config.WorkerParamPrefix}}.window_time`
- Variable-Name-Example: `analytics.window_time`
- Value: int (or string that can be unmarshalled to an integer)

### Input-IoT-Selection

- Desc: sets the iot selection of a flow-input-port
- Variable-Name-Template: `{{config.WorkerParamPrefix}}.selection.{{inputId}}.{{inputInPort}}`
- Variable-Name-Example: `analytics.selection.373808f2-848a-4446-8062-abd973dc96d3.value`
- Value: json.Marshal(model.IotOption{})
- Value-Example: `{"device_selection":{"device_id":"device_7","service_id":"s12","path":"root.value_s12.v2"}}`

### Input-IoT-Selection-Criteria

- Desc: if a selections does not contain a path (device-group selection), this parameter is needed to find one
- Variable-Name-Template: `{{config.WorkerParamPrefix}}.criteria.{{inputId}}.{{inputInPort}}`
- Variable-Name-Example: `analytics.criteria.373808f2-848a-4446-8062-abd973dc96d3.value`
- Value: json.Marshal([]devices.FilterCriteria{})
- Value-Example: `[{"function_id": "foo", "aspect_id": "bar"}]`

### Input-PersistData

- Desc: optional
- Variable-Name-Template: `{{config.WorkerParamPrefix}}.persistData.{{inputId}}`
- Variable-Name-Example: `analytics.persistData.373808f2-848a-4446-8062-abd973dc96d3`
- Value: json.Marshal(boolean)

### Input-Config

- Desc: sets the iot selection of a flow-input-port
- Variable-Name-Template: `{{config.WorkerParamPrefix}}.conf.{{inputId}}.{{inputConfigName}}`
- Variable-Name-Example: `analytics.conf.373808f2-848a-4446-8062-abd973dc96d3.url`
- Value: string
