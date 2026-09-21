# ShelfApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**deleteShelf**](ShelfApi.md#deleteshelf) | **DELETE** /v1/shelves/{shelfId} | Delete shelf |
| [**getPublicShelfByDomain**](ShelfApi.md#getpublicshelfbydomain) | **GET** /v1/shelves/by-domain/{domain} | Get public shelf by domain |
| [**getPublicShelfByPath**](ShelfApi.md#getpublicshelfbypath) | **GET** /v1/shelves/by-path/{path} | Get public shelf by path |
| [**getPublicShelfByUsernameAndPath**](ShelfApi.md#getpublicshelfbyusernameandpath) | **GET** /v1/shelves/by-user/{username}/{path} | Get public shelf by username and path |
| [**getShelfById**](ShelfApi.md#getshelfbyid) | **GET** /v1/shelves/{shelfId} | Get shelf by ID |
| [**listShelves**](ShelfApi.md#listshelves) | **GET** /v1/shelves | List shelves |
| [**postCreateShelf**](ShelfApi.md#postcreateshelf) | **POST** /v1/shelves | Create shelf |
| [**putUpdateShelf**](ShelfApi.md#putupdateshelf) | **PUT** /v1/shelves/{shelfId} | Update shelf |



## deleteShelf

> deleteShelf(shelfId)

Delete shelf

Delete a shelf by ID. Only its owner or an admin may delete it.

### Example

```ts
import {
  Configuration,
  ShelfApi,
} from '';
import type { DeleteShelfRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // Configure HTTP bearer authorization: bearer
    accessToken: "YOUR BEARER TOKEN",
  });
  const api = new ShelfApi(config);

  const body = {
    // string
    shelfId: shelfId_example,
  } satisfies DeleteShelfRequest;

  try {
    const data = await api.deleteShelf(body);
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
| **shelfId** | `string` |  | [Defaults to `undefined`] |

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


## getPublicShelfByDomain

> PublicShelf getPublicShelfByDomain(domain)

Get public shelf by domain

Get the public-safe view of the shelf that is served on a domain of its own, for example profile.example.com or profile.example.com:9443. The domain is normalized first (lowercased, without a trailing dot or slash, without :80 or :443). Used to render that shelf when the frontend is reached on the domain, and requires no authentication. Unlike the lookups by path it doesn\&#39;t depend on app.userBasedPaths.

### Example

```ts
import {
  Configuration,
  ShelfApi,
} from '';
import type { GetPublicShelfByDomainRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new ShelfApi();

  const body = {
    // string
    domain: domain_example,
  } satisfies GetPublicShelfByDomainRequest;

  try {
    const data = await api.getPublicShelfByDomain(body);
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
| **domain** | `string` |  | [Defaults to `undefined`] |

### Return type

[**PublicShelf**](PublicShelf.md)

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


## getPublicShelfByPath

> PublicShelf getPublicShelfByPath(path)

Get public shelf by path

Get the public-safe view of a shelf by its path. Used to render a shelf\&#39;s public link page and requires no authentication.

### Example

```ts
import {
  Configuration,
  ShelfApi,
} from '';
import type { GetPublicShelfByPathRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new ShelfApi();

  const body = {
    // string
    path: path_example,
  } satisfies GetPublicShelfByPathRequest;

  try {
    const data = await api.getPublicShelfByPath(body);
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
| **path** | `string` |  | [Defaults to `undefined`] |

### Return type

[**PublicShelf**](PublicShelf.md)

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


## getPublicShelfByUsernameAndPath

> PublicShelf getPublicShelfByUsernameAndPath(username, path)

Get public shelf by username and path

Get the public-safe view of a shelf by its owner\&#39;s username and its path, used while app.userBasedPaths is enabled. Requires no authentication. While the setting is disabled, use the lookup by path alone instead - the two never answer at the same time.

### Example

```ts
import {
  Configuration,
  ShelfApi,
} from '';
import type { GetPublicShelfByUsernameAndPathRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new ShelfApi();

  const body = {
    // string
    username: username_example,
    // string
    path: path_example,
  } satisfies GetPublicShelfByUsernameAndPathRequest;

  try {
    const data = await api.getPublicShelfByUsernameAndPath(body);
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
| **username** | `string` |  | [Defaults to `undefined`] |
| **path** | `string` |  | [Defaults to `undefined`] |

### Return type

[**PublicShelf**](PublicShelf.md)

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


## getShelfById

> Shelf getShelfById(shelfId)

Get shelf by ID

Get a shelf by ID. Only its owner or an admin may access it.

### Example

```ts
import {
  Configuration,
  ShelfApi,
} from '';
import type { GetShelfByIdRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // Configure HTTP bearer authorization: bearer
    accessToken: "YOUR BEARER TOKEN",
  });
  const api = new ShelfApi(config);

  const body = {
    // string
    shelfId: shelfId_example,
  } satisfies GetShelfByIdRequest;

  try {
    const data = await api.getShelfById(body);
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
| **shelfId** | `string` |  | [Defaults to `undefined`] |

### Return type

[**Shelf**](Shelf.md)

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


## listShelves

> Array&lt;Shelf&gt; listShelves()

List shelves

List shelves - the caller\&#39;s own, or every shelf for an admin.

### Example

```ts
import {
  Configuration,
  ShelfApi,
} from '';
import type { ListShelvesRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // Configure HTTP bearer authorization: bearer
    accessToken: "YOUR BEARER TOKEN",
  });
  const api = new ShelfApi(config);

  try {
    const data = await api.listShelves();
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

[**Array&lt;Shelf&gt;**](Shelf.md)

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


## postCreateShelf

> Shelf postCreateShelf(shelfBase)

Create shelf

Create a new shelf, owned by the authenticated caller.

### Example

```ts
import {
  Configuration,
  ShelfApi,
} from '';
import type { PostCreateShelfRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // Configure HTTP bearer authorization: bearer
    accessToken: "YOUR BEARER TOKEN",
  });
  const api = new ShelfApi(config);

  const body = {
    // ShelfBase
    shelfBase: ...,
  } satisfies PostCreateShelfRequest;

  try {
    const data = await api.postCreateShelf(body);
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
| **shelfBase** | [ShelfBase](ShelfBase.md) |  | |

### Return type

[**Shelf**](Shelf.md)

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


## putUpdateShelf

> Shelf putUpdateShelf(shelfId, shelfBase)

Update shelf

Update an existing shelf. Only its owner or an admin may update it.

### Example

```ts
import {
  Configuration,
  ShelfApi,
} from '';
import type { PutUpdateShelfRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // Configure HTTP bearer authorization: bearer
    accessToken: "YOUR BEARER TOKEN",
  });
  const api = new ShelfApi(config);

  const body = {
    // string
    shelfId: shelfId_example,
    // ShelfBase
    shelfBase: ...,
  } satisfies PutUpdateShelfRequest;

  try {
    const data = await api.putUpdateShelf(body);
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
| **shelfId** | `string` |  | [Defaults to `undefined`] |
| **shelfBase** | [ShelfBase](ShelfBase.md) |  | |

### Return type

[**Shelf**](Shelf.md)

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

