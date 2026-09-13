# ThemeApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**deleteTheme**](ThemeApi.md#deletetheme) | **DELETE** /v1/themes/{themeId} | Delete theme |
| [**getThemeById**](ThemeApi.md#getthemebyid) | **GET** /v1/themes/{themeId} | Get theme by ID |
| [**listAllUserThemes**](ThemeApi.md#listalluserthemes) | **GET** /v1/themes/admin | List all user themes |
| [**listThemes**](ThemeApi.md#listthemes) | **GET** /v1/themes | List themes |
| [**postCreateTheme**](ThemeApi.md#postcreatetheme) | **POST** /v1/themes | Create theme |
| [**putUpdateTheme**](ThemeApi.md#putupdatetheme) | **PUT** /v1/themes/{themeId} | Update theme |



## deleteTheme

> deleteTheme(themeId)

Delete theme

Delete a theme by ID. Its owner or an admin may delete a user theme; instance themes can never be deleted through the API.

### Example

```ts
import {
  Configuration,
  ThemeApi,
} from '';
import type { DeleteThemeRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // Configure HTTP bearer authorization: bearer
    accessToken: "YOUR BEARER TOKEN",
  });
  const api = new ThemeApi(config);

  const body = {
    // string
    themeId: themeId_example,
  } satisfies DeleteThemeRequest;

  try {
    const data = await api.deleteTheme(body);
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
| **themeId** | `string` |  | [Defaults to `undefined`] |

### Return type

`void` (Empty response body)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/problem+json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | No Content |  -  |
| **0** | Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## getThemeById

> Theme getThemeById(themeId)

Get theme by ID

Get a theme by ID. Instance themes are readable by any authenticated caller; user themes only by their owner or an admin.

### Example

```ts
import {
  Configuration,
  ThemeApi,
} from '';
import type { GetThemeByIdRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // Configure HTTP bearer authorization: bearer
    accessToken: "YOUR BEARER TOKEN",
  });
  const api = new ThemeApi(config);

  const body = {
    // string
    themeId: themeId_example,
  } satisfies GetThemeByIdRequest;

  try {
    const data = await api.getThemeById(body);
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
| **themeId** | `string` |  | [Defaults to `undefined`] |

### Return type

[**Theme**](Theme.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`, `application/problem+json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **0** | Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## listAllUserThemes

> Array&lt;Theme&gt; listAllUserThemes()

List all user themes

List every user-created theme across the instance, for moderation. Admin only.

### Example

```ts
import {
  Configuration,
  ThemeApi,
} from '';
import type { ListAllUserThemesRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // Configure HTTP bearer authorization: bearer
    accessToken: "YOUR BEARER TOKEN",
  });
  const api = new ThemeApi(config);

  try {
    const data = await api.listAllUserThemes();
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**Array&lt;Theme&gt;**](Theme.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`, `application/problem+json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **0** | Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## listThemes

> ThemeGroupedResponseBody listThemes()

List themes

List themes available to the caller, grouped into instance-provided and their own.

### Example

```ts
import {
  Configuration,
  ThemeApi,
} from '';
import type { ListThemesRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // Configure HTTP bearer authorization: bearer
    accessToken: "YOUR BEARER TOKEN",
  });
  const api = new ThemeApi(config);

  try {
    const data = await api.listThemes();
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**ThemeGroupedResponseBody**](ThemeGroupedResponseBody.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`, `application/problem+json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **0** | Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## postCreateTheme

> Theme postCreateTheme(themeBase)

Create theme

Create a new theme, owned by the authenticated caller.

### Example

```ts
import {
  Configuration,
  ThemeApi,
} from '';
import type { PostCreateThemeRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // Configure HTTP bearer authorization: bearer
    accessToken: "YOUR BEARER TOKEN",
  });
  const api = new ThemeApi(config);

  const body = {
    // ThemeBase
    themeBase: ...,
  } satisfies PostCreateThemeRequest;

  try {
    const data = await api.postCreateTheme(body);
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
| **themeBase** | [ThemeBase](ThemeBase.md) |  | |

### Return type

[**Theme**](Theme.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`, `application/problem+json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Created |  -  |
| **0** | Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## putUpdateTheme

> Theme putUpdateTheme(themeId, themeBase)

Update theme

Update an existing theme. Only its owner may update it - not even an admin, and never an instance theme (managed via the mounted directory instead).

### Example

```ts
import {
  Configuration,
  ThemeApi,
} from '';
import type { PutUpdateThemeRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // Configure HTTP bearer authorization: bearer
    accessToken: "YOUR BEARER TOKEN",
  });
  const api = new ThemeApi(config);

  const body = {
    // string
    themeId: themeId_example,
    // ThemeBase
    themeBase: ...,
  } satisfies PutUpdateThemeRequest;

  try {
    const data = await api.putUpdateTheme(body);
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
| **themeId** | `string` |  | [Defaults to `undefined`] |
| **themeBase** | [ThemeBase](ThemeBase.md) |  | |

### Return type

[**Theme**](Theme.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`, `application/problem+json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **0** | Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)

