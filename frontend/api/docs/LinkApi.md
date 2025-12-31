# LinkApi

All URIs are relative to *http://localhost:8085*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**deleteLink**](LinkApi.md#deletelink) | **DELETE** /v1/links/{linkId} | Delete link |
| [**getLinks**](LinkApi.md#getlinks) | **GET** /v1/links | Get links by shelf ID |
| [**postCreateLink**](LinkApi.md#postcreatelink) | **POST** /v1/links | Create link |
| [**putUpdateLink**](LinkApi.md#putupdatelink) | **PUT** /v1/links/{linkId} | Update link |



## deleteLink

> deleteLink(linkId, shelfId)

Delete link

Delete a link by ID.

### Example

```ts
import {
  Configuration,
  LinkApi,
} from '';
import type { DeleteLinkRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new LinkApi();

  const body = {
    // string
    linkId: linkId_example,
    // string (optional)
    shelfId: shelfId_example,
  } satisfies DeleteLinkRequest;

  try {
    const data = await api.deleteLink(body);
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
| **linkId** | `string` |  | [Defaults to `undefined`] |
| **shelfId** | `string` |  | [Optional] [Defaults to `undefined`] |

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


## getLinks

> Array&lt;Link&gt; getLinks(shelfId)

Get links by shelf ID

Get links by shelf ID.

### Example

```ts
import {
  Configuration,
  LinkApi,
} from '';
import type { GetLinksRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new LinkApi();

  const body = {
    // string (optional)
    shelfId: shelfId_example,
  } satisfies GetLinksRequest;

  try {
    const data = await api.getLinks(body);
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

[**Array&lt;Link&gt;**](Link.md)

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


## postCreateLink

> Link postCreateLink(linkBase)

Create link

Create a new link.

### Example

```ts
import {
  Configuration,
  LinkApi,
} from '';
import type { PostCreateLinkRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new LinkApi();

  const body = {
    // LinkBase
    linkBase: ...,
  } satisfies PostCreateLinkRequest;

  try {
    const data = await api.postCreateLink(body);
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
| **linkBase** | [LinkBase](LinkBase.md) |  | |

### Return type

[**Link**](Link.md)

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


## putUpdateLink

> Link putUpdateLink(linkId, linkBase, shelfId)

Update link

Update an existing link.

### Example

```ts
import {
  Configuration,
  LinkApi,
} from '';
import type { PutUpdateLinkRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new LinkApi();

  const body = {
    // string
    linkId: linkId_example,
    // LinkBase
    linkBase: ...,
    // string (optional)
    shelfId: shelfId_example,
  } satisfies PutUpdateLinkRequest;

  try {
    const data = await api.putUpdateLink(body);
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
| **linkId** | `string` |  | [Defaults to `undefined`] |
| **linkBase** | [LinkBase](LinkBase.md) |  | |
| **shelfId** | `string` |  | [Optional] [Defaults to `undefined`] |

### Return type

[**Link**](Link.md)

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

