
# LinkOrderResponseBody


## Properties

Name | Type
------------ | -------------
`$schema` | string
`failures` | [Array&lt;LinkOrderFailure&gt;](LinkOrderFailure.md)

## Example

```typescript
import type { LinkOrderResponseBody } from ''

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "failures": null,
} satisfies LinkOrderResponseBody

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as LinkOrderResponseBody
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


