package web

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"text/template"
)

var (
	funcMap = template.FuncMap{
		"pad": func(totalWidth int, padValues ...any) string {
			var stringIn string = ""
			var padString string = " "

			if len(padValues) > 0 {
				stringIn = fmt.Sprintf(`%v`, padValues[0])
				if len(padValues) > 1 {
					padString = fmt.Sprintf(`%v`, padValues[1])
				}
			}

			if len(stringIn) >= totalWidth {
				return stringIn
			}
			paddingLength := totalWidth - len(stringIn)
			leftPad := paddingLength >> 1
			if leftPad < 1 {
				return stringIn
			}
			if paddingLength-leftPad < 1 {
				return strings.Repeat(padString, leftPad) + stringIn
			}
			return strings.Repeat(padString, leftPad) + stringIn + strings.Repeat(padString, paddingLength-leftPad)
		},
		"lpad": func(totalWidth int, padValues ...any) string {
			var stringIn string = ""
			var padString string = " "

			if len(padValues) > 0 {
				stringIn = fmt.Sprintf(`%v`, padValues[0])
				if len(padValues) > 1 {
					padString = fmt.Sprintf(`%v`, padValues[1])
				}
			}

			if len(stringIn) >= totalWidth {
				return stringIn
			}
			paddingLength := totalWidth - len(stringIn)
			if paddingLength < 1 {
				return stringIn
			}
			return strings.Repeat(padString, paddingLength) + stringIn
		},
		"rpad": func(totalWidth int, padValues ...any) string {
			var stringIn string = ""
			var padString string = " "

			if len(padValues) > 0 {
				stringIn = fmt.Sprintf(`%v`, padValues[0])
				if len(padValues) > 1 {
					padString = fmt.Sprintf(`%v`, padValues[1])
				}
			}

			if len(stringIn) >= totalWidth {
				return stringIn
			}
			paddingLength := totalWidth - len(stringIn)
			if paddingLength < 1 {
				return stringIn
			}
			return stringIn + strings.Repeat(padString, paddingLength)
		},
		"join": func(items []string, sep string) string { return strings.Join(items, sep) },
		"lte":  func(a, b int) bool { return a <= b },
		"gte":  func(a, b int) bool { return a >= b },
		"lt":   func(a, b int) bool { return a < b },
		//"gt":   func(a, b int) bool { return a > b },
		"uc": func(s string) string { return strings.Title(s) },
		"lc": func(s string) string { return strings.ToLower(s) },
	}
)

func adminIndex(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("index.html").Funcs(funcMap).ParseFiles("web/html/admin/_header.html", "web/html/admin/index.html", "web/html/admin/_footer.html")
	if err != nil {
		slog.Error("HTML ERROR", "error", err)
	}

	tmpl.Execute(w, nil)

}
