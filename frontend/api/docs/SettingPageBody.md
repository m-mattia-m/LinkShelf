
# SettingPageBody


## Properties

Name | Type
------------ | -------------
`$schema` | string
`about` | string
`aboutShow` | boolean
`contact` | string
`contactShow` | boolean
`imprint` | string
`imprintShow` | boolean
`privacyPolicy` | string
`privacyPolicyShow` | boolean
`redirectToDashboard` | boolean
`termsOfUse` | string
`termsOfUseShow` | boolean

## Example

```typescript
import type { SettingPageBody } from ''

// TODO: Update the object below with actual values
const example = {
  "$schema": null,
  "about": null,
  "aboutShow": null,
  "contact": null,
  "contactShow": null,
  "imprint": null,
  "imprintShow": null,
  "privacyPolicy": null,
  "privacyPolicyShow": null,
  "redirectToDashboard": null,
  "termsOfUse": null,
  "termsOfUseShow": null,
} satisfies SettingPageBody

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as SettingPageBody
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


