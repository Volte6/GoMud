package i18n

import (
	"errors"
	"path"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

var (
	ErrMessageFallback = errors.New("i18n message fallback to default language")
)

var i18 *I18n

type BundleCfg struct {
	DefaultLanguage language.Tag
	Language        language.Tag
	LanguagePaths   []string
}

type I18n struct {
	bundle          *i18n.Bundle
	localizerByLng  map[language.Tag]*i18n.Localizer
	defaultLanguage language.Tag
}

func Init(c BundleCfg) {
	i18 = New(c)
}

func New(c BundleCfg) *I18n {
	i := &I18n{}

	bundle := i18n.NewBundle(c.DefaultLanguage)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	i.bundle = bundle
	i.defaultLanguage = c.DefaultLanguage
	i.localizerByLng = map[language.Tag]*i18n.Localizer{}

	for _, p := range c.LanguagePaths {
		bundle.LoadMessageFile(path.Join(p, c.DefaultLanguage.String()+".yaml"))
		if c.DefaultLanguage != c.Language {
			bundle.LoadMessageFile(path.Join(p, c.Language.String()+".yaml"))
		}
	}

	i.localizerByLng[c.Language] = i.newLocalizer(c.Language)

	// set defaultLanguage if it isn't exist
	if _, hasDefaultLng := i.localizerByLng[i.defaultLanguage]; !hasDefaultLng {
		i.localizerByLng[i.defaultLanguage] = i.newLocalizer(i.defaultLanguage)
	}

	return i
}

func (i *I18n) newLocalizer(lng language.Tag) *i18n.Localizer {
	lngDefault := i.defaultLanguage.String()
	lngs := []string{
		lng.String(),
	}

	if lng.String() != lngDefault {
		lngs = append(lngs, lngDefault)
	}

	localizer := i18n.NewLocalizer(
		i.bundle,
		lngs...,
	)

	return localizer
}

// Translate message.
func Translate(lng language.Tag, msgID string, tplData ...map[any]any) (string, error) {
	return TranslateWithConfig(i18, lng, msgID, tplData...)
}

// Translate message.
func TranslateWithConfig(i *I18n, lng language.Tag, msgID string, tplData ...map[any]any) (string, error) {
	localizer, ok := i.localizerByLng[lng]
	if !ok {
		localizer = i.localizerByLng[i.defaultLanguage]
	}

	cfg := &i18n.LocalizeConfig{
		MessageID: msgID,
	}

	if len(tplData) > 0 && tplData[0] != nil {
		cfg.TemplateData = tplData[0]
	}

	msg, l, err := localizer.LocalizeWithTag(cfg)
	if err != nil {
		// Fallback to English
		if !l.IsRoot() {
			return msg, ErrMessageFallback
		}

		// Fallback to language.Und
		return msgID, err
	}

	if l != lng {
		return msg, ErrMessageFallback
	}

	return msg, nil
}

func IsMessageNotFoundErr(err error) bool {
	_, ok := err.(*i18n.MessageNotFoundErr)

	return ok
}

func IsMessageFallbackErr(err error) bool {
	return errors.Is(err, ErrMessageFallback)
}
