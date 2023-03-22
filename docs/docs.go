// Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "#soon",
        "contact": {
            "name": "Burak Fidan",
            "email": "mrntlu@gmail.com"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/anime/season": {
            "get": {
                "description": "Returns animes by year and season",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "anime"
                ],
                "summary": "Get Animes by Year and Season",
                "parameters": [
                    {
                        "description": "Sort Anime By Year and Season",
                        "name": "sortbyyearseasonanime",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/requests.SortByYearSeasonAnime"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/responses.Anime"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/anime/upcoming": {
            "get": {
                "description": "Returns upcoming animes by sort",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "anime"
                ],
                "summary": "Get Upcoming Animes by Sort",
                "parameters": [
                    {
                        "description": "Sort Anime",
                        "name": "sortanime",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/requests.SortAnime"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/responses.Anime"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/auth/register": {
            "post": {
                "description": "Allows users to register",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "User Registration",
                "parameters": [
                    {
                        "description": "User registration info",
                        "name": "register",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/requests.Register"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/user": {
            "delete": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Deletes everything related to user",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Deletes user information",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Authentication header",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/user/forgot-password": {
            "post": {
                "description": "Users can change their password when they forgot",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Will be used when user forgot password",
                "parameters": [
                    {
                        "description": "User's email",
                        "name": "ForgotPassword",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/requests.ForgotPassword"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Couldn't find any user",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/user/info": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Returns users membership \u0026 investing/subscription limits",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "User membership info",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Authentication header",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "User Info",
                        "schema": {
                            "$ref": "#/definitions/responses.UserInfo"
                        }
                    }
                }
            }
        },
        "/user/membership": {
            "patch": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "User membership status will be updated depending on subscription status",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Change User Membership",
                "parameters": [
                    {
                        "description": "Set Membership",
                        "name": "changemembership",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/requests.ChangeMembership"
                        }
                    },
                    {
                        "type": "string",
                        "description": "Authentication header",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/user/notification": {
            "patch": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Users can change their notification preference",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Change User Notification Preference",
                "parameters": [
                    {
                        "description": "Set notification",
                        "name": "changenotification",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/requests.ChangeNotification"
                        }
                    },
                    {
                        "type": "string",
                        "description": "Authentication header",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/user/password": {
            "patch": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Users can change their password",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Change User Password",
                "parameters": [
                    {
                        "description": "Set new password",
                        "name": "ChangePassword",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/requests.ChangePassword"
                        }
                    },
                    {
                        "type": "string",
                        "description": "Authentication header",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/user/token": {
            "patch": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Depending on logged in device fcm token will be updated",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Updates FCM User Token",
                "parameters": [
                    {
                        "description": "Set token",
                        "name": "changefcmtoken",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/requests.ChangeFCMToken"
                        }
                    },
                    {
                        "type": "string",
                        "description": "Authentication header",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "requests.ChangeFCMToken": {
            "type": "object",
            "required": [
                "fcm_token"
            ],
            "properties": {
                "fcm_token": {
                    "type": "string"
                }
            }
        },
        "requests.ChangeMembership": {
            "type": "object",
            "properties": {
                "is_premium": {
                    "type": "boolean"
                }
            }
        },
        "requests.ChangeNotification": {
            "type": "object",
            "required": [
                "app_notification",
                "mail_notification"
            ],
            "properties": {
                "app_notification": {
                    "type": "boolean"
                },
                "mail_notification": {
                    "type": "boolean"
                }
            }
        },
        "requests.ChangePassword": {
            "type": "object",
            "required": [
                "new_password",
                "old_password"
            ],
            "properties": {
                "new_password": {
                    "type": "string",
                    "minLength": 6
                },
                "old_password": {
                    "type": "string",
                    "minLength": 6
                }
            }
        },
        "requests.ForgotPassword": {
            "type": "object",
            "required": [
                "email_address"
            ],
            "properties": {
                "email_address": {
                    "type": "string"
                }
            }
        },
        "requests.Register": {
            "type": "object",
            "required": [
                "email_address",
                "password",
                "username"
            ],
            "properties": {
                "email_address": {
                    "type": "string"
                },
                "fcm_token": {
                    "type": "string"
                },
                "image": {
                    "type": "string"
                },
                "password": {
                    "type": "string",
                    "minLength": 6
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "requests.SortAnime": {
            "type": "object",
            "required": [
                "page",
                "sort",
                "type"
            ],
            "properties": {
                "page": {
                    "type": "integer",
                    "minimum": 1
                },
                "sort": {
                    "type": "string",
                    "enum": [
                        "popularity",
                        "date"
                    ]
                },
                "type": {
                    "type": "integer",
                    "enum": [
                        1,
                        -1
                    ]
                }
            }
        },
        "requests.SortByYearSeasonAnime": {
            "type": "object",
            "required": [
                "page",
                "season",
                "sort",
                "type",
                "year"
            ],
            "properties": {
                "page": {
                    "type": "integer",
                    "minimum": 1
                },
                "season": {
                    "type": "string",
                    "enum": [
                        "winter",
                        "summer",
                        "fall",
                        "spring"
                    ]
                },
                "sort": {
                    "type": "string",
                    "enum": [
                        "popularity",
                        "date"
                    ]
                },
                "type": {
                    "type": "integer",
                    "enum": [
                        1,
                        -1
                    ]
                },
                "year": {
                    "type": "integer"
                }
            }
        },
        "responses.Anime": {
            "type": "object",
            "properties": {
                "age_rating": {
                    "type": "string"
                },
                "aired": {
                    "$ref": "#/definitions/responses.AnimeAirDate"
                },
                "demographics": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/responses.AnimeGenre"
                    }
                },
                "description": {
                    "type": "string"
                },
                "episodes": {
                    "type": "integer"
                },
                "genres": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/responses.AnimeGenre"
                    }
                },
                "image_url": {
                    "type": "string"
                },
                "is_airing": {
                    "type": "boolean"
                },
                "mal_id": {
                    "type": "integer"
                },
                "mal_score": {
                    "type": "number"
                },
                "mal_scored_by": {
                    "type": "integer"
                },
                "producers": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/responses.AnimeNameURL"
                    }
                },
                "relations": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/responses.AnimeRelation"
                    }
                },
                "season": {
                    "type": "string"
                },
                "small_image_url": {
                    "type": "string"
                },
                "source": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                },
                "streaming": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/responses.AnimeNameURL"
                    }
                },
                "studios": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/responses.AnimeNameURL"
                    }
                },
                "themes": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/responses.AnimeGenre"
                    }
                },
                "title_en": {
                    "type": "string"
                },
                "title_jp": {
                    "type": "string"
                },
                "title_original": {
                    "type": "string"
                },
                "trailer": {
                    "type": "string"
                },
                "type": {
                    "type": "string"
                },
                "year": {
                    "type": "integer"
                }
            }
        },
        "responses.AnimeAirDate": {
            "type": "object",
            "properties": {
                "from": {
                    "type": "string"
                },
                "from_day": {
                    "type": "integer"
                },
                "from_month": {
                    "type": "integer"
                },
                "from_year": {
                    "type": "integer"
                },
                "to": {
                    "type": "string"
                },
                "to_day": {
                    "type": "integer"
                },
                "to_month": {
                    "type": "integer"
                },
                "to_year": {
                    "type": "integer"
                }
            }
        },
        "responses.AnimeGenre": {
            "type": "object",
            "properties": {
                "mal_id": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                },
                "url": {
                    "type": "string"
                }
            }
        },
        "responses.AnimeNameURL": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string"
                },
                "url": {
                    "type": "string"
                }
            }
        },
        "responses.AnimeRelation": {
            "type": "object",
            "properties": {
                "relation": {
                    "type": "string"
                },
                "source": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/responses.AnimeRelationDetails"
                    }
                }
            }
        },
        "responses.AnimeRelationDetails": {
            "type": "object",
            "properties": {
                "mal_id": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                },
                "redirect_url": {
                    "type": "string"
                },
                "type": {
                    "type": "string"
                }
            }
        },
        "responses.UserInfo": {
            "type": "object",
            "properties": {
                "app_notification": {
                    "type": "boolean"
                },
                "email": {
                    "type": "string"
                },
                "fcm_token": {
                    "type": "string"
                },
                "is_oauth": {
                    "type": "boolean"
                },
                "is_premium": {
                    "type": "boolean"
                },
                "mail_notification": {
                    "type": "boolean"
                },
                "username": {
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "http://localhost:8080",
	BasePath:         "/api/v1",
	Schemes:          []string{"https"},
	Title:            "Project Consumer API",
	Description:      "REST Api of Project Consumer.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
