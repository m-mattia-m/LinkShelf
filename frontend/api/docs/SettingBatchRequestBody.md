
# SettingBatchRequestBody


## Properties

Name | Type
------------ | -------------
`$schema` | string
`settings` | [Array&lt;Setting&gt;](Setting.md)

## Example

```typescript
import type { SettingBatchRequestBody } from ''

// TODO: Update the object below with actual values
const example = {
  "$schema": http://localhost:8085/schemas/SettingBatchRequestBody.json,
  "settings": null,
} satisfies SettingBatchRequestBody

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as SettingBatchRequestBody
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


