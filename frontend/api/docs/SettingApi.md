# SettingApi

All URIs are relative to *http://localhost:8085*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**getEmailDeliveryInfo**](SettingApi.md#getemaildeliveryinfo) | **GET** /v1/settings/email-delivery | Get email delivery info |
| [**getPageSettings**](SettingApi.md#getpagesettings) | **GET** /v1/settings | Get page settings |
| [**putUpdateSetting**](SettingApi.md#putupdatesetting) | **PUT** /v1/settings | Update setting |
| [**putUpdateSettingsBatch**](SettingApi.md#putupdatesettingsbatch) | **PUT** /v1/settings/batch | Update settings in batch |



## getEmailDeliveryInfo

> EmailDeliveryInfo getEmailDeliveryInfo()

Get email delivery info

Admin-only, read-only: the SMTP host and from-address currently configured. SMTP itself is configured exclusively via config, not through this API.

### Example

```ts
import {
  Configuration,
  SettingApi,
} from '';
import type { GetEmailDeliveryInfoRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // Configure HTTP bearer authorization: bearer
    accessToken: "YOUR BEARER TOKEN",
  });
  const api = new SettingApi(config);

  try {
    const data = await api.getEmailDeliveryInfo();
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

[**EmailDeliveryInfo**](EmailDeliveryInfo.md)

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


## getPageSettings

> SettingPageBody getPageSettings(languageCode)

Get page settings

Get page settings by language code. Used to render the public site shell (title, contact info, legal pages, ...) and requires no authentication.

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
  const config = new Configuration({ 
    // Configure HTTP bearer authorization: bearer
    accessToken: "YOUR BEARER TOKEN",
  });
  const api = new SettingApi(config);

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


## putUpdateSettingsBatch

> SettingBatchResponseBody putUpdateSettingsBatch(settingBatchRequestBody)

Update settings in batch

Update many page settings in a single request. Invalid items are rejected individually (reported in the response\&#39;s failures list) without aborting the rest of the batch.

### Example

```ts
import {
  Configuration,
  SettingApi,
} from '';
import type { PutUpdateSettingsBatchRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // Configure HTTP bearer authorization: bearer
    accessToken: "YOUR BEARER TOKEN",
  });
  const api = new SettingApi(config);

  const body = {
    // SettingBatchRequestBody
    settingBatchRequestBody: ...,
  } satisfies PutUpdateSettingsBatchRequest;

  try {
    const data = await api.putUpdateSettingsBatch(body);
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
| **settingBatchRequestBody** | [SettingBatchRequestBody](SettingBatchRequestBody.md) |  | |

### Return type

[**SettingBatchResponseBody**](SettingBatchResponseBody.md)

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

