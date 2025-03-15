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
			name: "not exist msgID",
			args: args{
				lng:   language.German,
				msgID: "notExistMsgID",
			},
			want:    "notExistMsgID",
			errFunc: IsMessageNotFoundErr,
		},
		// French (fallback)
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
			name: "not exist msgID",
			args: args{
				lng:   language.German,
				msgID: "notExistMsgID",
			},
			want:    "notExistMsgID",
			errFunc: IsMessageNotFoundErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.args.lng.String()+" - "+tt.args.msgID+" - "+tt.name, func(t *testing.T) {
			i18 := New(BundleCfg{
				DefaultLanguage: language.English,
				Language:        tt.args.lng,
				LanguagePaths:   []string{"testdata/localize"},
			})

			got, err := TranslateWithConfig(i18, tt.args.lng, tt.args.msgID, tt.args.tplData)
			if got != tt.want {
				t.Errorf("TranslateWithConfig() = %v, want %v", got, tt.want)
			}

			if tt.errFunc == nil {
				if err != nil {
					t.Errorf("TranslateWithConfig() unexpected error: %v", err)
				}
			} else {
				if !tt.errFunc(err) {
					t.Errorf("TranslateWithConfig() unexpected error: %v", err)
				}
			}
		})
	}
}
