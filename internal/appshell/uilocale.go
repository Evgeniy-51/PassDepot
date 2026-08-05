package appshell

import (
	"fmt"
	"sync/atomic"
)

// uiLang — язык сообщений для UI: "ru" | "en".
var uiLang atomic.Value

func init() {
	uiLang.Store("ru")
}

// SetUILocale задаёт язык ошибок и системных диалогов (вызывается из UI при смене языка).
func (a *App) SetUILocale(locale string) {
	if locale != "en" {
		locale = "ru"
	}
	uiLang.Store(locale)
}

func currentUILang() string {
	if v, ok := uiLang.Load().(string); ok {
		return v
	}
	return "ru"
}

// L возвращает ru или en строку в зависимости от текущего языка UI.
func L(ru, en string) string {
	if currentUILang() == "en" {
		return en
	}
	return ru
}

func Lf(ruFmt, enFmt string, args ...any) string {
	if currentUILang() == "en" {
		return fmt.Sprintf(enFmt, args...)
	}
	return fmt.Sprintf(ruFmt, args...)
}
