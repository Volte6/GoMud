package language

import (
	"errors"
	"path"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

var (
	ErrMessageFallback = errors.New("translation message fallback to default language")
)

var trans *Translation

type BundleCfg struct {
	DefaultLanguage language.Tag
	Language        language.Tag
	LanguagePaths   []string
}

type Translation struct {
	bundle          *i18n.Bundle
	localizerByLng  map[language.Tag]*i18n.Localizer
	defaultLanguage language.Tag
}

func InitTranslation(c BundleCfg) {
	trans = NewTranslation(c)
}

func NewTranslation(c BundleCfg) *Translation {
	t := &Translation{}

	bundle := i18n.NewBundle(c.DefaultLanguage)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	t.bundle = bundle
	t.defaultLanguage = c.DefaultLanguage
	t.localizerByLng = map[language.Tag]*i18n.Localizer{}

	for _, p := range c.LanguagePaths {
		bundle.LoadMessageFile(path.Join(p, c.DefaultLanguage.String()+".yaml"))
		if c.DefaultLanguage != c.Language {
			bundle.LoadMessageFile(path.Join(p, c.Language.String()+".yaml"))
		}
	}

	t.localizerByLng[c.Language] = t.newLocalizer(c.Language)

	// Add defaultLanguage if it isn't exist
	if _, hasDefaultLng := t.localizerByLng[t.defaultLanguage]; !hasDefaultLng {
		t.localizerByLng[t.defaultLanguage] = t.newLocalizer(t.defaultLanguage)
	}

	return t
}

func (t *Translation) newLocalizer(lng language.Tag) *i18n.Localizer {
	lngDefault := t.defaultLanguage.String()
	lngs := []string{
		lng.String(),
	}

	if lng.String() != lngDefault {
		lngs = append(lngs, lngDefault)
	}

	localizer := i18n.NewLocalizer(
		t.bundle,
		lngs...,
	)

	return localizer
}

func T(msgID string, tplData ...map[any]any) string {
	lng := language.Make(configs.GetTranslationConfig().Language.String())

	msg, err := trans.Translate(lng, msgID, tplData...)
	if err != nil {
		if !IsMessageNotFoundErr(err) && !IsMessageFallbackErr(err) {
			mudlog.Error(`Translation`, "msgID", msgID, `error`, err)
		}
	}

	return msg
}

// Translate message.
func (t *Translation) Translate(lng language.Tag, msgID string, tplData ...map[any]any) (string, error) {
	localizer, ok := t.localizerByLng[lng]
	if !ok {
		localizer = t.localizerByLng[t.defaultLanguage]
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
