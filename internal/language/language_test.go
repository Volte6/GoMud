package language

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

func TestTranslate(t *testing.T) {
	type args struct {
		lng     language.Tag
		msgID   string
		tplData map[any]any
	}
	tests := []struct {
		name    string
		args    args
		want    string
		errFunc func(error) bool
	}{
		{
			name: "hello",
			args: args{
				lng:   language.English,
				msgID: "welcome",
			},
			want: "hello",
		},
		{
			name: "hello alex",
			args: args{
				lng:   language.English,
				msgID: "welcomeWithName",
				tplData: map[any]any{
					"name": "alex",
				},
			},
			want: "hello alex",
		},
		{
			name: "18 years old",
			args: args{
				lng:   language.English,
				msgID: "welcomeWithAge",
				tplData: map[any]any{
					"age": "18",
				},
			},
			want: "I am 18 years old",
		},
		{
			name: "fallback to english",
			args: args{
				lng:   language.English,
				msgID: "fallbackToEnglish",
			},
			want: "fallback to english",
		},
		{
			name: "fallback to english2",
			args: args{
				lng:   language.English,
				msgID: "fallbackToEnglish2",
			},
			want: "fallback to english2",
		},
		{
			name: "fallback to msgID",
			args: args{
				lng:   language.English,
				msgID: "fallbackToMsgID",
			},
			want:    "fallbackToMsgID",
			errFunc: IsMessageNotFoundErr,
		},
		{
			name: "not exist msgID",
			args: args{
				lng:   language.English,
				msgID: "notExistMsgID",
			},
			want:    "notExistMsgID",
			errFunc: IsMessageNotFoundErr,
		},
		// German
		{
			name: "hallo",
			args: args{
				lng:   language.German,
				msgID: "welcome",
			},
			want: "hallo",
		},
		{
			name: "hallo alex",
			args: args{
				lng:   language.German,
				msgID: "welcomeWithName",
				tplData: map[any]any{
					"name": "alex",
				},
			},
			want: "hallo alex",
		},
		{
			name: "18 jahre alt",
			args: args{
				lng:   language.German,
				msgID: "welcomeWithAge",
				tplData: map[any]any{
					"age": "18",
				},
			},
			want: "ich bin 18 Jahre alt",
		},
		{
			name: "fallback to english",
			args: args{
				lng:   language.German,
				msgID: "fallbackToEnglish",
			},
			want:    "fallback to english",
			errFunc: IsMessageFallbackErr,
		},
		{
			name: "fallback to english2",
			args: args{
				lng:   language.German,
				msgID: "fallbackToEnglish2",
			},
			want:    "fallback to english2",
			errFunc: IsMessageFallbackErr,
		},
		{
			name: "fallback to msgID",
			args: args{
				lng:   language.German,
				msgID: "fallbackToMsgID",
			},
			want:    "fallbackToMsgID",
			errFunc: IsMessageNotFoundErr,
		},
		{
			name: "not exist msgID",
			args: args{
				lng:   language.German,
				msgID: "notExistMsgID",
			},
			want:    "notExistMsgID",
			errFunc: IsMessageNotFoundErr,
		},
		// French (not exist language fallback)
		{
			name: "hello",
			args: args{
				lng:   language.French,
				msgID: "welcome",
			},
			want:    "hello",
			errFunc: IsMessageFallbackErr,
		},
		{
			name: "hello alex",
			args: args{
				lng:   language.French,
				msgID: "welcomeWithName",
				tplData: map[any]any{
					"name": "alex",
				},
			},
			want:    "hello alex",
			errFunc: IsMessageFallbackErr,
		},
		{
			name: "18 years old",
			args: args{
				lng:   language.French,
				msgID: "welcomeWithAge",
				tplData: map[any]any{
					"age": "18",
				},
			},
			want:    "I am 18 years old",
			errFunc: IsMessageFallbackErr,
		},
		{
			name: "fallback to english",
			args: args{
				lng:   language.French,
				msgID: "fallbackToEnglish",
			},
			want:    "fallback to english",
			errFunc: IsMessageFallbackErr,
		},
		{
			name: "fallback to english2",
			args: args{
				lng:   language.French,
				msgID: "fallbackToEnglish2",
			},
			want:    "fallback to english2",
			errFunc: IsMessageFallbackErr,
		},
		{
			name: "fallback to msgID",
			args: args{
				lng:   language.French,
				msgID: "fallbackToMsgID",
			},
			want:    "fallbackToMsgID",
			errFunc: IsMessageNotFoundErr,
		},
		{
			name: "not exist msgID",
			args: args{
				lng:   language.French,
				msgID: "notExistMsgID",
			},
			want:    "notExistMsgID",
			errFunc: IsMessageNotFoundErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.args.lng.String()+" - "+tt.args.msgID+" - "+tt.name, func(t *testing.T) {
			trans := NewTranslation(BundleCfg{
				DefaultLanguage: language.English,
				Language:        tt.args.lng,
				LanguagePaths:   []string{"testdata/localize"},
			})

			got, err := trans.Translate(tt.args.lng, tt.args.msgID, tt.args.tplData)
			if got != tt.want {
				t.Errorf("Translate() = %v, want %v", got, tt.want)
			}

			if tt.errFunc == nil {
				if err != nil {
					t.Errorf("Translate(),  msgID = %v, unexpected error: %v", tt.args.msgID, err)
				}
			} else {
				if !tt.errFunc(err) {
					t.Errorf("Translate(),  msgID = %v, unexpected error: %v", tt.args.msgID, err)
				}
			}
		})
	}
}

func TestTranslateReload(t *testing.T) {
	// Initial
	args := map[string]string{
		"welcome": "hallo",
	}
	bs, _ := yaml.Marshal(args)
	os.WriteFile("testdata/reload-test/de.yaml", bs, 0644)

	trans := NewTranslation(BundleCfg{
		DefaultLanguage: language.English,
		Language:        language.German,
		LanguagePaths:   []string{"testdata/reload-test"},
	})

	got, _ := trans.Translate(language.German, "welcome")
	assert.Equal(t, "hallo", got)
	got, _ = trans.Translate(language.German, "welcomeWithName")
	assert.Equal(t, "welcomeWithName", got)

	// First modification (update & add)
	args = map[string]string{
		"welcome":         "willkommen",
		"welcomeWithName": "hallo {{ .name }}",
	}
	bs, _ = yaml.Marshal(args)
	os.WriteFile("testdata/reload-test/de.yaml", bs, 0644)

	trans.ReloadTranslation()

	got, _ = trans.Translate(language.German, "welcome")
	assert.Equal(t, "willkommen", got)
	got, _ = trans.Translate(language.German, "welcomeWithName", map[any]any{"name": "alex"})
	assert.Equal(t, "hallo alex", got)

	// Second modification (update & delete)
	args = map[string]string{
		"welcomeWithName": "willkommen {{ .name }}",
	}
	bs, _ = yaml.Marshal(args)
	os.WriteFile("testdata/reload-test/de.yaml", bs, 0644)

	trans.ReloadTranslation()

	got, _ = trans.Translate(language.German, "welcome")
	assert.Equal(t, "welcome", got)
	got, _ = trans.Translate(language.German, "welcomeWithName", map[any]any{"name": "alex"})
	assert.Equal(t, "willkommen alex", got)
}
