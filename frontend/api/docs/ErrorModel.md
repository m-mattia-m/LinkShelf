
# ErrorModel


## Properties

Name | Type
------------ | -------------
`$schema` | string
`detail` | string
`errors` | [Array&lt;ErrorDetail&gt;](ErrorDetail.md)
`instance` | string
`status` | number
`title` | string
`type` | string

## Example

```typescript
import type { ErrorModel } from ''

// TODO: Update the object below with actual values
const example = {
  "$schema": http://localhost:8085/schemas/ErrorModel.json,
  "detail": Property foo is required but is missing.,
  "errors": null,
  "instance": https://example.com/error-log/abc123,
  "status": 400,
  "title": Bad Request,
  "type": https://example.com/errors/example,
} satisfies ErrorModel

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as ErrorModel
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


