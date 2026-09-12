
# ResendVerificationRequest


## Properties

Name | Type
------------ | -------------
`$schema` | string
`email` | string

## Example

```typescript
import type { ResendVerificationRequest } from ''

// TODO: Update the object below with actual values
const example = {
  "$schema": http://localhost:8085/schemas/ResendVerificationRequest.json,
  "email": null,
} satisfies ResendVerificationRequest

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as ResendVerificationRequest
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


