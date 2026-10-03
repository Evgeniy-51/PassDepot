package appshell

import (
	"errors"
	"os"
	"strings"
)

// formatPullError превращает ошибку git/файла при подтягивании remote в короткое сообщение для UI.
func formatPullError(err error) string {
	if err == nil {
		return ""
	}
	low := strings.ToLower(err.Error())
	switch {
	case strings.Contains(low, "timeout"), strings.Contains(low, "i/o timeout"), strings.Contains(low, "wsarecv"):
		return L("Не удалось связаться с GitHub (GitLab) (таймаут сети).", "Failed to reach GitHub (GitLab): network timeout.")
	case strings.Contains(low, "could not resolve host"), strings.Contains(low, "no such host"):
		return L("Не удалось установить соединение (проверьте сеть и DNS).", "Failed to establish connection (check network and DNS).")
	case strings.Contains(low, "connection refused"), strings.Contains(low, "connection reset"):
		return L("Соединение прервано или отклонено.", "Connection was reset or refused.")
	case strings.Contains(low, "authentication failed"), strings.Contains(low, "access denied"),
		strings.Contains(low, "could not read username"),
		strings.Contains(low, "invalid username or token"), strings.Contains(low, "could not authenticate"):
		return msgAccessDenied()
	case strings.Contains(low, "repository not found"):
		return L("Репозиторий не найден или нет доступа.", "Repository not found or access denied.")
	case strings.Contains(low, "401"), strings.Contains(low, "403"):
		return msgAccessDenied()
	}
	var perr *os.PathError
	if errors.As(err, &perr) && errors.Is(perr.Err, os.ErrNotExist) {
		return L("Файл базы не найден после обновления (нет в клоне или неверный путь).", "Vault file not found after refresh (missing from clone or invalid path).")
	}
	if errors.Is(err, os.ErrNotExist) {
		return L("Файл базы не найден после обновления (нет в клоне или неверный путь).", "Vault file not found after refresh (missing from clone or invalid path).")
	}
	return Lf("Обновление из репозитория: %s", "Repository refresh failed: %s", err.Error())
}

func msgAccessDenied() string {
	return L("Отказ доступа к репозиторию: проверьте PAT и права.", "Repository access denied: check the PAT and permissions.")
}

// isAuthError — ошибка git, которую formatPullError показывает как отказ доступа (неверный/отозванный PAT).
func isAuthError(err error) bool {
	return err != nil && formatPullError(err) == msgAccessDenied()
}

// needPATPrefix помечает ошибку входа, после которой UI должен запросить PAT.
// Фронт ищет префикс и убирает его из текста (App.tsx, parseNeedPat).
const needPATPrefix = "[need-pat] "

func needPATError(msg string) error {
	return errors.New(needPATPrefix + msg)
}
