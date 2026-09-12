# AuthApi

All URIs are relative to *http://localhost:8085*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**getOidcLogin**](AuthApi.md#getoidclogin) | **GET** /v1/auth/oidc/login | Start OIDC login |
| [**postLogin**](AuthApi.md#postlogin) | **POST** /v1/auth/login | Login |
| [**postLogout**](AuthApi.md#postlogout) | **POST** /v1/auth/logout | Logout |
| [**postOidcCallback**](AuthApi.md#postoidccallback) | **POST** /v1/auth/oidc/callback | Complete OIDC login |
| [**postRefresh**](AuthApi.md#postrefresh) | **POST** /v1/auth/refresh | Refresh token |
| [**postResendVerification**](AuthApi.md#postresendverification) | **POST** /v1/auth/resend-verification | Resend verification email |
| [**postSetPassword**](AuthApi.md#postsetpassword) | **POST** /v1/auth/set-password | Set password |
| [**postVerifyEmail**](AuthApi.md#postverifyemail) | **POST** /v1/auth/verify-email | Verify email |



## getOidcLogin

> OidcLoginResponseBody getOidcLogin()

Start OIDC login

Returns the authorization URL (PKCE-protected) the frontend should redirect the browser to, and the state to send back to the callback.

### Example

```ts
import {
  Configuration,
  AuthApi,
} from '';
import type { GetOidcLoginRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new AuthApi();

  try {
    const data = await api.getOidcLogin();
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

[**OidcLoginResponseBody**](OidcLoginResponseBody.md)

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


## postLogin

> TokenPair postLogin(loginRequest)

Login

Log in with a local username and password, returning an access/refresh token pair.

### Example

```ts
import {
  Configuration,
  AuthApi,
} from '';
import type { PostLoginRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new AuthApi();

  const body = {
    // LoginRequest
    loginRequest: ...,
  } satisfies PostLoginRequest;

  try {
    const data = await api.postLogin(body);
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
| **loginRequest** | [LoginRequest](LoginRequest.md) |  | |

### Return type

[**TokenPair**](TokenPair.md)

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


## postLogout

> postLogout(refreshRequest)

Logout

Invalidate a refresh token. The current access token simply expires on its own shortly after.

### Example

```ts
import {
  Configuration,
  AuthApi,
} from '';
import type { PostLogoutRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new AuthApi();

  const body = {
    // RefreshRequest
    refreshRequest: ...,
  } satisfies PostLogoutRequest;

  try {
    const data = await api.postLogout(body);
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
| **refreshRequest** | [RefreshRequest](RefreshRequest.md) |  | |

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


## postOidcCallback

> TokenPair postOidcCallback(oidcCallbackRequest, authorization)

Complete OIDC login

Exchanges an authorization code for tokens. Behaves as login-or-auto-provision when called anonymously, or as link-to-my-account when called with a valid Bearer token.

### Example

```ts
import {
  Configuration,
  AuthApi,
} from '';
import type { PostOidcCallbackRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new AuthApi();

  const body = {
    // OidcCallbackRequest
    oidcCallbackRequest: ...,
    // string | Optional bearer token. When present and valid, the external identity is linked to the already-authenticated user instead of logging in as a new/matched user. (optional)
    authorization: authorization_example,
  } satisfies PostOidcCallbackRequest;

  try {
    const data = await api.postOidcCallback(body);
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
| **oidcCallbackRequest** | [OidcCallbackRequest](OidcCallbackRequest.md) |  | |
| **authorization** | `string` | Optional bearer token. When present and valid, the external identity is linked to the already-authenticated user instead of logging in as a new/matched user. | [Optional] [Defaults to `undefined`] |

### Return type

[**TokenPair**](TokenPair.md)

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


## postRefresh

> TokenPair postRefresh(refreshRequest)

Refresh token

Exchange a refresh token for a new access/refresh token pair. The refresh token used is invalidated (single-use rotation).

### Example

```ts
import {
  Configuration,
  AuthApi,
} from '';
import type { PostRefreshRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new AuthApi();

  const body = {
    // RefreshRequest
    refreshRequest: ...,
  } satisfies PostRefreshRequest;

  try {
    const data = await api.postRefresh(body);
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
| **refreshRequest** | [RefreshRequest](RefreshRequest.md) |  | |

### Return type

[**TokenPair**](TokenPair.md)

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


## postResendVerification

> postResendVerification(resendVerificationRequest)

Resend verification email

Re-sends whatever verification/invite link is still pending for the given email, rate-limited. Always responds the same way regardless of whether the address exists, is already verified, or was rate-limited.

### Example

```ts
import {
  Configuration,
  AuthApi,
} from '';
import type { PostResendVerificationRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new AuthApi();

  const body = {
    // ResendVerificationRequest
    resendVerificationRequest: ...,
  } satisfies PostResendVerificationRequest;

  try {
    const data = await api.postResendVerification(body);
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
| **resendVerificationRequest** | [ResendVerificationRequest](ResendVerificationRequest.md) |  | |

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


## postSetPassword

> postSetPassword(setPasswordRequest)

Set password

Completes the admin-invite flow: sets an account\&#39;s first password and marks it verified, using the token from the emailed link.

### Example

```ts
import {
  Configuration,
  AuthApi,
} from '';
import type { PostSetPasswordRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new AuthApi();

  const body = {
    // SetPasswordRequest
    setPasswordRequest: ...,
  } satisfies PostSetPasswordRequest;

  try {
    const data = await api.postSetPassword(body);
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
| **setPasswordRequest** | [SetPasswordRequest](SetPasswordRequest.md) |  | |

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


## postVerifyEmail

> postVerifyEmail(verifyEmailRequest)

Verify email

Completes email verification for an account that already has a password, using the token from the emailed link.

### Example

```ts
import {
  Configuration,
  AuthApi,
} from '';
import type { PostVerifyEmailRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new AuthApi();

  const body = {
    // VerifyEmailRequest
    verifyEmailRequest: ...,
  } satisfies PostVerifyEmailRequest;

  try {
    const data = await api.postVerifyEmail(body);
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
| **verifyEmailRequest** | [VerifyEmailRequest](VerifyEmailRequest.md) |  | |

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

