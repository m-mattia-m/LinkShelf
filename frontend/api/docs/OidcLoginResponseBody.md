
# OidcLoginResponseBody


## Properties

Name | Type
------------ | -------------
`$schema` | string
`authorizationUrl` | string
`state` | string

## Example

```typescript
import type { OidcLoginResponseBody } from ''

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "authorizationUrl": null,
  "state": null,
} satisfies OidcLoginResponseBody

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as OidcLoginResponseBody
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


