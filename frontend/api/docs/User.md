
# User


## Properties

Name | Type
------------ | -------------
`$schema` | string
`email` | string
`firstName` | string
`id` | string
`lastName` | string
`password` | string

## Example

```typescript
import type { User } from ''

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "email": null,
  "firstName": null,
  "id": null,
  "lastName": null,
  "password": null,
} satisfies User

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as User
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


