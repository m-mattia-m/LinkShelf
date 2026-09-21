
# Shelf


## Properties

Name | Type
------------ | -------------
`$schema` | string
`createdWithUserBasedPaths` | boolean
`description` | string
`domain` | string
`icon` | string
`id` | string
`path` | string
`theme` | { [key: string]: string; }
`themeId` | string
`themeMissing` | boolean
`title` | string
`userId` | string
`username` | string

## Example

```typescript
import type { Shelf } from ''

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "createdWithUserBasedPaths": null,
  "description": null,
  "domain": null,
  "icon": null,
  "id": null,
  "path": null,
  "theme": null,
  "themeId": null,
  "themeMissing": null,
  "title": null,
  "userId": null,
  "username": null,
} satisfies Shelf

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as Shelf
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


