# SettingApi

All URIs are relative to *http://localhost:8085*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**getPageSettings**](SettingApi.md#getpagesettings) | **GET** /v1/settings | Get page settings |
| [**putUpdateSetting**](SettingApi.md#putupdatesetting) | **PUT** /v1/settings | Update setting |



## getPageSettings

> SettingPageBody getPageSettings(languageCode)

Get page settings

Get page settings by language code.

### Example

```ts
import {
  Configuration,
  SettingApi,
} from '';
import type { GetPageSettingsRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new SettingApi();

  const body = {
    // string (optional)
    languageCode: languageCode_example,
  } satisfies GetPageSettingsRequest;

  try {
    const data = await api.getPageSettings(body);
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
| **languageCode** | `string` |  | [Optional] [Defaults to `undefined`] |

### Return type

[**SettingPageBody**](SettingPageBody.md)

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


## putUpdateSetting

> SettingPageBody putUpdateSetting(setting)

Update setting

Update page settings.

### Example

```ts
import {
  Configuration,
  SettingApi,
} from '';
import type { PutUpdateSettingRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new SettingApi();

  const body = {
    // Setting
    setting: ...,
  } satisfies PutUpdateSettingRequest;

  try {
    const data = await api.putUpdateSetting(body);
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
| **setting** | [Setting](Setting.md) |  | |

### Return type

[**SettingPageBody**](SettingPageBody.md)

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

