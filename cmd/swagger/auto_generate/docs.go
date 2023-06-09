// Code generated by swaggo/swag. DO NOT EDIT.

package auto_generate

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/tasks": {
            "get": {
                "description": "查询任务列表，可根据状态、类型和优先级过滤",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Tasks"
                ],
                "summary": "查询任务列表",
                "parameters": [
                    {
                        "type": "string",
                        "description": "任务状态",
                        "name": "status",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "任务类型",
                        "name": "type",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "任务优先级",
                        "name": "priority",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.ResponseProtocol"
                        }
                    }
                }
            },
            "post": {
                "description": "提交任务到Scheduler",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Tasks"
                ],
                "summary": "提交任务",
                "parameters": [
                    {
                        "description": "任务对象",
                        "name": "task",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handler.SubmitTaskRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.ResponseProtocol"
                        }
                    }
                }
            }
        },
        "/api/tasks/run_again/{taskID}": {
            "post": {
                "description": "将指定任务再次运行，此过程会返回新的 taskID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Tasks"
                ],
                "summary": "重跑任务",
                "parameters": [
                    {
                        "type": "string",
                        "description": "任务ID",
                        "name": "taskID",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.ResponseProtocol"
                        }
                    }
                }
            }
        },
        "/api/tasks/{taskID}": {
            "get": {
                "description": "查询特定任务的执行状态",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Tasks"
                ],
                "summary": "查询任务状态",
                "parameters": [
                    {
                        "type": "string",
                        "description": "任务ID",
                        "name": "taskID",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.ResponseProtocol"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "errors.ErrorCode": {
            "type": "integer",
            "enum": [
                0,
                400,
                401,
                402,
                403,
                500
            ],
            "x-enum-varnames": [
                "OK",
                "BadRequest",
                "Unauthorized",
                "Forbidden",
                "ResourceNotExist",
                "InternalServerError"
            ]
        },
        "handler.ResponseProtocol": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "business code, 0 means ok",
                    "allOf": [
                        {
                            "$ref": "#/definitions/errors.ErrorCode"
                        }
                    ]
                },
                "data": {
                    "description": "result"
                },
                "message": {
                    "description": "error message",
                    "type": "string"
                }
            }
        },
        "handler.SubmitTaskRequest": {
            "type": "object",
            "required": [
                "kind",
                "priority",
                "task"
            ],
            "properties": {
                "kind": {
                    "type": "string"
                },
                "priority": {
                    "type": "string",
                    "enum": [
                        "high",
                        "medium",
                        "low"
                    ]
                },
                "task": {}
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
