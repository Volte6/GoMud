# Translation

## Configuration

```yaml
Translation:
  # - DefaultLanguage -
  # Specify the default game language (fallback)
  DefaultLanguage: 'en'
  # - Language -
  # Specify the game language
  Language: 'en'
  # - LanguagePaths -
  # Specify the game language file paths
  LanguagePaths:
    - '_datafiles/localize'
    - '_datafiles/world/default/localize'
```

## Localize Files Format

All localize files are in `LanguagePaths`, filename format is: `<language>.yaml`, language can be found at: https://github.com/golang/text/blob/master/language/tags.go

Localize files example:

* English: `en.yaml`
* German: `de.yaml`
* Chinese: `zh.yaml`

Localize file content format is: `msgID: 'Translation value'`

Example:

```yaml
MsgFoo: 'Message Foo'
'Msg Bar': 'Message Bar'
Hello:
```

> Note: If the `Translation value` is equals to the `msgID`, the `Translation value` can be set to empty

### Translation Fallback

1. If the `Translation value` of a `msgID` for the current language does not exist or is empty, it will fallback to the default language (usually English).
2. If the `Translation value` of a `msgID` in the default language does not exist or is empty, it will fall back to `msgID`.

## Translation and Usage Example

### Invoke Translate Function in Golang

```golang
import (
	"github.com/volte6/gomud/internal/language"
)

// Simple message
language.T(`Hello`)

// Message with arguments
language.T(`Hello {{ .name }}`, map[string]string{
	"name": "alex",
})

// Message with string format
fmt.Sprintf(language.T(`%d users online`), userCt)
```

### Invoke Translate Function in Templates

```golang
import (
	"github.com/volte6/gomud/internal/lauguage"
)

// Passing arguments to template
templates.Process("foo/bar", map[any]any{
	"name": "alex",
})
```

#### Template with Translation Function

```golang
// Simple message
{{ t "Hello" }}

// Message with arguments
{{- $map := map "name" .name -}}
{{ t "Hello {{ .name }}" $map }}

// Message with string format
{{ t "%d users online" }}
```

### Localize Files

`en.yaml`:

```yaml
Hello: 'Hello'
'Hello {{ .name }}': 'Hello {{ .name }}'
'%d users online': '%d users online'
```

`de.yaml`

```yaml
Hello: 'Hallo'
'Hello {{ .name }}': 'Hallo {{ .name }}'
'%d users online': '%d benutzer online'
```

### Translation with `templates.ProcessText()`

If the text to be translated is neither in Golang nor in a template, such as from the configuration file, you need to invoke `templates.ProcessText()` to translate it.

Example of motd in configuration:

```golang
m := configs.GetServerConfig().Motd.String()
text, err := templates.ProcessText(m, nil)
if err != nil {
    text = m
}

user.SendText(text)
```

## TODO

* Completion of all texts to be translated and localized files
* Other scenarios that require translation functionality, such as room descriptions, item names/descriptions, etc.
* Idea: Specify different port numbers (Telnet/HTTP) to output different languages
