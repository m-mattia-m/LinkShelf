-- Appended, not replaced, so an operator who already rewrote their privacy
-- policy keeps their own wording - this only adds to it. Worded
-- conditionally ("may use... if enabled") since this seeds every instance
-- regardless of whether that instance actually turns Plausible on.
UPDATE `setting`
SET `value` = CONCAT(`value`, '

## Analytics

This website may use [Plausible Analytics](https://plausible.io), a privacy-friendly, cookie-free analytics tool, if enabled by the site operator. Plausible does not use cookies and does not collect personally identifiable information. See the [Plausible data policy](https://plausible.io/data-policy) for details.
')
WHERE `key` = 'privacy_policy' AND `language` = 'en';
