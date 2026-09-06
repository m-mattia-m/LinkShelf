# SectionApi

All URIs are relative to *http://localhost:8085*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**deleteSection**](SectionApi.md#deletesection) | **DELETE** /v1/sections/{sectionId} | Delete section |
| [**getSections**](SectionApi.md#getsections) | **GET** /v1/sections | Get sections by shelf ID |
| [**postCreateSection**](SectionApi.md#postcreatesection) | **POST** /v1/sections | Create section |
| [**putUpdateSection**](SectionApi.md#putupdatesection) | **PUT** /v1/sections/{sectionId} | Update section |



## deleteSection

> deleteSection(sectionId)

Delete section

Delete a section by ID. Only the owner of its shelf or an admin may delete it.

### Example

```ts
import {
  Configuration,
  SectionApi,
} from '';
import type { DeleteSectionRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // Configure HTTP bearer authorization: bearer
    accessToken: "YOUR BEARER TOKEN",
  });
  const api = new SectionApi(config);

  const body = {
    // string
    sectionId: sectionId_example,
  } satisfies DeleteSectionRequest;

  try {
    const data = await api.deleteSection(body);
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
| **sectionId** | `string` |  | [Defaults to `undefined`] |

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


## getSections

> Array&lt;Section&gt; getSections(shelfId)

Get sections by shelf ID

Get sections by shelf ID. Used to render a shelf\&#39;s public link page and requires no authentication.

### Example

```ts
import {
  Configuration,
  SectionApi,
} from '';
import type { GetSectionsRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new SectionApi();

  const body = {
    // string (optional)
    shelfId: shelfId_example,
  } satisfies GetSectionsRequest;

  try {
    const data = await api.getSections(body);
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
| **shelfId** | `string` |  | [Optional] [Defaults to `undefined`] |

### Return type

[**Array&lt;Section&gt;**](Section.md)

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


## postCreateSection

> Section postCreateSection(sectionBase)

Create section

Create a new section on a shelf owned by the caller.

### Example

```ts
import {
  Configuration,
  SectionApi,
} from '';
import type { PostCreateSectionRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // Configure HTTP bearer authorization: bearer
    accessToken: "YOUR BEARER TOKEN",
  });
  const api = new SectionApi(config);

  const body = {
    // SectionBase
    sectionBase: ...,
  } satisfies PostCreateSectionRequest;

  try {
    const data = await api.postCreateSection(body);
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
| **sectionBase** | [SectionBase](SectionBase.md) |  | |

### Return type

[**Section**](Section.md)

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


## putUpdateSection

> Section putUpdateSection(sectionId, sectionBase)

Update section

Update an existing section. Only the owner of its shelf or an admin may update it.

### Example

```ts
import {
  Configuration,
  SectionApi,
} from '';
import type { PutUpdateSectionRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // Configure HTTP bearer authorization: bearer
    accessToken: "YOUR BEARER TOKEN",
  });
  const api = new SectionApi(config);

  const body = {
    // string
    sectionId: sectionId_example,
    // SectionBase
    sectionBase: ...,
  } satisfies PutUpdateSectionRequest;

  try {
    const data = await api.putUpdateSection(body);
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
| **sectionId** | `string` |  | [Defaults to `undefined`] |
| **sectionBase** | [SectionBase](SectionBase.md) |  | |

### Return type

[**Section**](Section.md)

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

