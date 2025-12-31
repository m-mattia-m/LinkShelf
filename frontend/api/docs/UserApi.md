# UserApi

All URIs are relative to *http://localhost:8085*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**deleteUser**](UserApi.md#deleteuser) | **DELETE** /v1/users/{userId} | Delete user |
| [**getUserById**](UserApi.md#getuserbyid) | **GET** /v1/users/{userId} | Get user by ID |
| [**patchUserPassword**](UserApi.md#patchuserpassword) | **PATCH** /v1/users/{userId}/password | Patch user password |
| [**postCreateUser**](UserApi.md#postcreateuser) | **POST** /v1/users | Create user |
| [**putUpdateUser**](UserApi.md#putupdateuser) | **PUT** /v1/users/{userId} | Update user |



## deleteUser

> deleteUser(userId)

Delete user

Delete a user by ID.

### Example

```ts
import {
  Configuration,
  UserApi,
} from '';
import type { DeleteUserRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new UserApi();

  const body = {
    // string | The identifier of the chosen form you want.
    userId: userId_example,
  } satisfies DeleteUserRequest;

  try {
    const data = await api.deleteUser(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **userId** | `string` | The identifier of the chosen form you want. | [Defaults to `undefined`] |

### Return type

`void` (Empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/problem+json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | No Content |  -  |
| **0** | Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## getUserById

> User getUserById(userId)

Get user by ID

Get a user by ID.

### Example

```ts
import {
  Configuration,
  UserApi,
} from '';
import type { GetUserByIdRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new UserApi();

  const body = {
    // string | The identifier of the chosen form you want.
    userId: userId_example,
  } satisfies GetUserByIdRequest;

  try {
    const data = await api.getUserById(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **userId** | `string` | The identifier of the chosen form you want. | [Defaults to `undefined`] |

### Return type

[**User**](User.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`, `application/problem+json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **0** | Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## patchUserPassword

> patchUserPassword(userId, userRequestBodyOnlyPassword)

Patch user password

Patch a user\&#39;s password.

### Example

```ts
import {
  Configuration,
  UserApi,
} from '';
import type { PatchUserPasswordRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new UserApi();

  const body = {
    // string | The identifier of the chosen form you want.
    userId: userId_example,
    // UserRequestBodyOnlyPassword
    userRequestBodyOnlyPassword: ...,
  } satisfies PatchUserPasswordRequest;

  try {
    const data = await api.patchUserPassword(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **userId** | `string` | The identifier of the chosen form you want. | [Defaults to `undefined`] |
| **userRequestBodyOnlyPassword** | [UserRequestBodyOnlyPassword](UserRequestBodyOnlyPassword.md) |  | |

### Return type

`void` (Empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/problem+json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | No Content |  -  |
| **0** | Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## postCreateUser

> User postCreateUser(userBase)

Create user

Create a new user.

### Example

```ts
import {
  Configuration,
  UserApi,
} from '';
import type { PostCreateUserRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new UserApi();

  const body = {
    // UserBase
    userBase: ...,
  } satisfies PostCreateUserRequest;

  try {
    const data = await api.postCreateUser(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **userBase** | [UserBase](UserBase.md) |  | |

### Return type

[**User**](User.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`, `application/problem+json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Created |  -  |
| **0** | Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## putUpdateUser

> User putUpdateUser(userId, userBase)

Update user

Update an existing user. Consider that password updates are not handled here.

### Example

```ts
import {
  Configuration,
  UserApi,
} from '';
import type { PutUpdateUserRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new UserApi();

  const body = {
    // string | The identifier of the chosen form you want.
    userId: userId_example,
    // UserBase
    userBase: ...,
  } satisfies PutUpdateUserRequest;

  try {
    const data = await api.putUpdateUser(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **userId** | `string` | The identifier of the chosen form you want. | [Defaults to `undefined`] |
| **userBase** | [UserBase](UserBase.md) |  | |

### Return type

[**User**](User.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`, `application/problem+json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **0** | Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)

