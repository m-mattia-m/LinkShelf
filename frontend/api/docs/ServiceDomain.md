
# ServiceDomain


## Properties

Name | Type
------------ | -------------
`$schema` | string
`createdAt` | Date
`domain` | string
`ownerId` | string
`shelfId` | string
`status` | [ServiceDomainStatus](ServiceDomainStatus.md)
`updatedAt` | Date

## Example

```typescript
import type { ServiceDomain } from ''

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "createdAt": null,
  "domain": null,
  "ownerId": null,
  "shelfId": null,
  "status": null,
  "updatedAt": null,
} satisfies ServiceDomain

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as ServiceDomain
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


