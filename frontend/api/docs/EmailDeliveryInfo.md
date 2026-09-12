
# EmailDeliveryInfo


## Properties

Name | Type
------------ | -------------
`$schema` | string
`enabled` | boolean
`from` | string
`host` | string

## Example

```typescript
import type { EmailDeliveryInfo } from ''

// TODO: Update the object below with actual values
const example = {
  "$schema": http://localhost:8085/schemas/EmailDeliveryInfo.json,
  "enabled": null,
  "from": null,
  "host": null,
} satisfies EmailDeliveryInfo

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as EmailDeliveryInfo
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


