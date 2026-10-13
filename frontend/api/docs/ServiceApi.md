# ServiceApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**listServiceDomains**](ServiceApi.md#listservicedomains) | **GET** /v1/service/domains | List custom domains |
| [**setServiceDomainStatus**](ServiceApi.md#setservicedomainstatus) | **PUT** /v1/service/domains/{domain}/status | Set the status of a custom domain |



## listServiceDomains

> ServiceDomainList listServiceDomains(cursor, limit, ifNoneMatch)

List custom domains

Every shelf domain with the status an external controller reported, ordered by shelf id and cursor paginated. Service token with the scope domains:read only, a user token is rejected. Supports If-None-Match, answering 304 when the page did not change.

### Example

```ts
import {
  Configuration,
  ServiceApi,
} from '';
import type { ListServiceDomainsRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // Configure HTTP bearer authorization: bearer
    accessToken: "YOUR BEARER TOKEN",
  });
  const api = new ServiceApi(config);

  const body = {
    // string | The next_cursor of the previous page. (optional)
    cursor: cursor_example,
    // number (optional)
    limit: 789,
    // string | The ETag of an earlier response. Answers 304 when nothing changed. (optional)
    ifNoneMatch: ifNoneMatch_example,
  } satisfies ListServiceDomainsRequest;

  try {
    const data = await api.listServiceDomains(body);
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
| **cursor** | `string` | The next_cursor of the previous page. | [Optional] [Defaults to `undefined`] |
| **limit** | `number` |  | [Optional] [Defaults to `100`] |
| **ifNoneMatch** | `string` | The ETag of an earlier response. Answers 304 when nothing changed. | [Optional] [Defaults to `undefined`] |

### Return type

[**ServiceDomainList**](ServiceDomainList.md)

### Authorization

[bearer](../README.md#bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`, `application/problem+json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  * ETag -  <br>  |
| **304** | Not Modified |  -  |
| **422** | Unprocessable Entity |  -  |
| **500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## setServiceDomainStatus

> ServiceDomain setServiceDomainStatus(domain, serviceDomainStatusBody)

Set the status of a custom domain

An external controller reports whether a domain is pending, ready or in error, and the DNS records the user has to set. Answers 409 when the shelf does not have the domain anymore, or when another shelf is already ready on it. Service token with the scope domains:write-status only.

### Example

```ts
import {
  Configuration,
  ServiceApi,
} from '';
import type { SetServiceDomainStatusRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // Configure HTTP bearer authorization: bearer
    accessToken: "YOUR BEARER TOKEN",
  });
  const api = new ServiceApi(config);

  const body = {
    // string
    domain: domain_example,
    // ServiceDomainStatusBody
    serviceDomainStatusBody: ...,
  } satisfies SetServiceDomainStatusRequest;

  try {
    const data = await api.setServiceDomainStatus(body);
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
| **serviceDomainStatusBody** | [ServiceDomainStatusBody](ServiceDomainStatusBody.md) |  | |

### Return type

[**ServiceDomain**](ServiceDomain.md)

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

