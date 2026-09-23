
# ShelfBase


## Properties

Name | Type
------------ | -------------
`$schema` | string
`description` | string
`domain` | string
`footerCustomText` | string
`footerEnabled` | boolean
`icon` | string
`noIndex` | boolean
`path` | string
`themeId` | string
`title` | string

## Example

```typescript
import type { ShelfBase } from ''

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "description": null,
  "domain": null,
  "footerCustomText": null,
  "footerEnabled": null,
  "icon": null,
  "noIndex": null,
  "path": null,
  "themeId": null,
  "title": null,
} satisfies ShelfBase

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as ShelfBase
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


