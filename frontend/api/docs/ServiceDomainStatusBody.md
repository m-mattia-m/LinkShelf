
# ServiceDomainStatusBody


## Properties

Name | Type
------------ | -------------
`$schema` | string
`message` | string
`reason` | string
`records` | [Array&lt;DomainRecord&gt;](DomainRecord.md)
`shelfId` | string
`state` | string

## Example

```typescript
import type { ServiceDomainStatusBody } from ''

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "message": null,
  "reason": null,
  "records": null,
  "shelfId": null,
  "state": null,
} satisfies ServiceDomainStatusBody

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as ServiceDomainStatusBody
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


