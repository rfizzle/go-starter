


# Application
  

## Informations

### Version

0.2.0

## Content negotiation

### URI Schemes
  * http

### Consumes
  * application/json

### Produces
  * application/json

## Access control

### Security Schemes

#### hasPermission



> **Type**: oauth2
>
> **Flow**: accessCode
>
> **Authorization URL**: https://example.com
>
> **Token URL**: https://example.com
      

##### Scopes

Name | Description
-----|-------------
auth:check | Check if the user is authenticated

### Security Requirements
  * hasPermission: deny

## All endpoints

###  auth

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/v1/auth/check | [auth check v1](#auth-check-v1) | Check if the user is authenticated |
| POST | /api/v1/auth/login | [auth login v1](#auth-login-v1) | Login a user |
| POST | /api/v1/auth/logout | [auth logout v1](#auth-logout-v1) | Logout the current user |
  


###  health

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/healthz/liveness | [health liveness](#health-liveness) | Liveness probe for kubernetes health check. Returns 200 if the service is alive. |
| GET | /api/healthz/readiness | [health readiness](#health-readiness) | Readiness probe |
  


## Paths

### <span id="auth-check-v1"></span> Check if the user is authenticated (*AuthCheckV1*)

```
GET /api/v1/auth/check
```

Check if the user is authenticated


#### Security Requirements
  * hasPermission: auth:check

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#auth-check-v1-200) | OK | success |  | [schema](#auth-check-v1-200-schema) |
| [401](#auth-check-v1-401) | Unauthorized | unauthorized |  | [schema](#auth-check-v1-401-schema) |

#### Responses


##### <span id="auth-check-v1-200"></span> 200 - success
Status: OK

###### <span id="auth-check-v1-200-schema"></span> Schema

##### <span id="auth-check-v1-401"></span> 401 - unauthorized
Status: Unauthorized

###### <span id="auth-check-v1-401-schema"></span> Schema

### <span id="auth-login-v1"></span> Login a user (*AuthLoginV1*)

```
POST /api/v1/auth/login
```

Authenticates a user from a username and password and returns a JWT in the response and inside a
signed cookie.


#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| body | `body` | [LoginRequest](#login-request) | `models.LoginRequest` | | ✓ | |  |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#auth-login-v1-200) | OK | success |  | [schema](#auth-login-v1-200-schema) |
| [400](#auth-login-v1-400) | Bad Request | Bad request |  | [schema](#auth-login-v1-400-schema) |
| [500](#auth-login-v1-500) | Internal Server Error | Server error |  | [schema](#auth-login-v1-500-schema) |

#### Responses


##### <span id="auth-login-v1-200"></span> 200 - success
Status: OK

###### <span id="auth-login-v1-200-schema"></span> Schema
   
  

[LoginRequest](#login-request)

##### <span id="auth-login-v1-400"></span> 400 - Bad request
Status: Bad Request

###### <span id="auth-login-v1-400-schema"></span> Schema
   
  

[FailureResponse](#failure-response)

##### <span id="auth-login-v1-500"></span> 500 - Server error
Status: Internal Server Error

###### <span id="auth-login-v1-500-schema"></span> Schema
   
  

[ErrorResponse](#error-response)

### <span id="auth-logout-v1"></span> Logout the current user (*AuthLogoutV1*)

```
POST /api/v1/auth/logout
```

Invalidates an authenticated user's session and cookie


#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#auth-logout-v1-200) | OK | success |  | [schema](#auth-logout-v1-200-schema) |

#### Responses


##### <span id="auth-logout-v1-200"></span> 200 - success
Status: OK

###### <span id="auth-logout-v1-200-schema"></span> Schema

### <span id="health-liveness"></span> Liveness probe for kubernetes health check. Returns 200 if the service is alive. (*HealthLiveness*)

```
GET /api/healthz/liveness
```

Liveness probe

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#health-liveness-200) | OK | success |  | [schema](#health-liveness-200-schema) |

#### Responses


##### <span id="health-liveness-200"></span> 200 - success
Status: OK

###### <span id="health-liveness-200-schema"></span> Schema

### <span id="health-readiness"></span> Readiness probe (*HealthReadiness*)

```
GET /api/healthz/readiness
```

Readiness probe for kubernetes health check. Returns 200 if the service is ready to serve requests. 
Returns 503 if the service is not ready to serve requests (starting up or shutting down).


#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#health-readiness-200) | OK | success |  | [schema](#health-readiness-200-schema) |
| [503](#health-readiness-503) | Service Unavailable | Not available |  | [schema](#health-readiness-503-schema) |

#### Responses


##### <span id="health-readiness-200"></span> 200 - success
Status: OK

###### <span id="health-readiness-200-schema"></span> Schema

##### <span id="health-readiness-503"></span> 503 - Not available
Status: Service Unavailable

###### <span id="health-readiness-503-schema"></span> Schema
   
  

[ErrorResponse](#error-response)

## Models

### <span id="error-response"></span> errorResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| code | integer| `int64` |  | |  |  |
| data | [interface{}](#interface)| `interface{}` |  | |  |  |
| message | string| `string` | ✓ | |  |  |
| status | string| `string` | ✓ | |  |  |



### <span id="failure-response"></span> failureResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [interface{}](#interface)| `interface{}` | ✓ | |  |  |
| status | string| `string` | ✓ | |  |  |



### <span id="login-request"></span> loginRequest


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| email | string| `string` | ✓ | | Email of user |  |
| password | string| `string` | ✓ | | Password of user |  |


