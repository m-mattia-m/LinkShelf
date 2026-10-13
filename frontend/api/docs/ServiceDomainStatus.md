
# ServiceDomainStatus


## Properties

Name | Type
------------ | -------------
`message` | string
`observedDomain` | string
`reason` | string
`records` | [Array&lt;DomainRecord&gt;](DomainRecord.md)
`state` | string

## Example

```typescript
import type { ServiceDomainStatus } from ''

// TODO: Update the object below with actual values
const example = {
  "message": null,
  "observedDomain": null,
  "reason": null,
  "records": null,
  "state": null,
} satisfies ServiceDomainStatus

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as ServiceDomainStatus
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


