package service

type j map[string]interface{}

var swaggerMap = j{
	"basePath": "/",
	"swagger":  "2.0",
	"info": j{
		"title":       "Cats API",
		"version":     "0.0.1",
		"description": "API for assertions",
	},
	"tags": []j{
		j{
			"name":        "assertions",
			"description": "Assertions API",
		},
	},
	"paths": j{
		"/assertions": j{
			"get": j{
				"tags": []string{
					"assertions",
				},
				"operationId": "getAssertions",
				"description": "Get all assertions for the current authenticated user.",
				"parameters":  []string{},
				"responses": j{
					"200": j{
						"schema": j{
							"type": "array",
							"items": j{
								"$ref": "#/definitions/CheckAssertions",
							},
							"description": "Assertions",
						},
					},
				},
			},
			"post": j{
				"tags": []string{
					"assertions",
				},
				"parameters": []j{
					j{
						"schema": j{
							"$ref": "#/definitions/CheckAssertions",
						},
						"in":          "body",
						"required":    true,
						"name":        "CheckAssertions",
						"description": "post assertions for a check.",
					},
				},
				"responses": j{
					"200": j{
						"schema": j{
							"$ref": "#/definitions/CheckAssertions",
						},
					},
				},
			},
		},
		"/assertions/{check_id}": j{
			"delete": j{
				"tags": []string{
					"assertions",
				},
				"description": "Delete assertions by check id",
				"parameters": []j{
					j{
						"type":        "string",
						"in":          "path",
						"required":    true,
						"name":        "check_id",
						"description": "Check ID",
					},
				},
				"responses": j{
					"default": j{
						"description": "",
					},
				},
			},
			"put": j{
				"tags": []string{
					"assertions",
				},
				"description": "Replace assertions for check.",
				"paramters": []j{
					j{
						"schema": j{
							"$ref": "#/definitions/CheckAssertions",
						},
						"in":          "body",
						"required":    true,
						"name":        "CheckAssertions",
						"description": "New CheckAssertions for the check.",
					},
					j{
						"type":        "string",
						"name":        "check_id",
						"in":          "path",
						"required":    true,
						"description": "Check ID",
					},
				},
				"responses": j{
					"200": j{
						"schema": j{
							"$ref": "#/definitions/CheckAssertions",
						},
					},
					"description": "",
				},
			},
			"get": j{
				"tags": []string{
					"assertions",
				},
				"description": "Get assertions for check.",
				"paramters": []j{
					j{
						"type":        "string",
						"name":        "check_id",
						"in":          "path",
						"required":    true,
						"description": "Check ID",
					},
				},
				"responses": j{
					"200": j{
						"schema": j{
							"$ref": "#/definitions/CheckAssertions",
						},
					},
					"description": "",
				},
			},
		},
	},
}
