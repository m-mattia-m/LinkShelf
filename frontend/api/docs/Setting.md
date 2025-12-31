
# Setting


## Properties

Name | Type
------------ | -------------
`$schema` | string
`key` | string
`languageCode` | string
`value` | string

## Example

```typescript
import type { Setting } from ''

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "key": null,
  "languageCode": null,
  "value": null,
} satisfies Setting

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as Setting
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


