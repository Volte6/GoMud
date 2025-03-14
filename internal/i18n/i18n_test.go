package i18n

import (
	"testing"

	"golang.org/x/text/language"
)

func TestTranslate(t *testing.T) {
	type args struct {
		lng     language.Tag
		msgID   string
		tplData map[any]any
	}
	tests := []struct {
		name string
		args args
		want string
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
		// French (fallback)
		{
			name: "hello",
			args: args{
				lng:   language.French,
				msgID: "welcome",
			},
			want: "hello",
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
			want: "hello alex",
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
			want: "I am 18 years old",
		},
	}

	for _, tt := range tests {
		t.Run(tt.args.lng.String()+" - "+tt.args.msgID+" - "+tt.name, func(t *testing.T) {
			i18 := New(BundleCfg{
				DefaultLanguage: language.English,
				Language:        tt.args.lng,
				LanguagePaths:   []string{"testdata/localize"},
			})

			if got := TranslateWithConfig(i18, tt.args.lng, tt.args.msgID, tt.args.tplData); got != tt.want {
				t.Errorf("Translate() = %v, want %v", got, tt.want)
			}
		})
	}
}
