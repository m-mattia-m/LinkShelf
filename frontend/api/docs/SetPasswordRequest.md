
# SetPasswordRequest


## Properties

Name | Type
------------ | -------------
`$schema` | string
`newPassword` | string
`token` | string

## Example

```typescript
import type { SetPasswordRequest } from ''

// TODO: Update the object below with actual values
const example = {
  "$schema": http://localhost:8085/schemas/SetPasswordRequest.json,
  "newPassword": null,
  "token": null,
} satisfies SetPasswordRequest

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as SetPasswordRequest
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


