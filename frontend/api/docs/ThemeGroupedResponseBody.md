
# ThemeGroupedResponseBody


## Properties

Name | Type
------------ | -------------
`$schema` | string
`instance` | [Array&lt;Theme&gt;](Theme.md)
`mine` | [Array&lt;Theme&gt;](Theme.md)

## Example

```typescript
import type { ThemeGroupedResponseBody } from ''

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "instance": null,
  "mine": null,
} satisfies ThemeGroupedResponseBody

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as ThemeGroupedResponseBody
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


