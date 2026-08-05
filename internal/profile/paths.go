package profile

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const appDirName = "PassDepot"

// DataRoot — каталог данных приложения: %AppData% (Roaming)\PassDepot.
// Для тестов: переменная окружения PASSDEPOT_DATA_ROOT.
func DataRoot() (string, error) {
	if d := os.Getenv("PASSDEPOT_DATA_ROOT"); d != "" {
		return filepath.Abs(d)
	}
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfg, appDirName), nil
}

// EnsureDataRoot создаёт корень данных при отсутствии.
func EnsureDataRoot() error {
	root, err := DataRoot()
	if err != nil {
		return err
	}
	return os.MkdirAll(root, 0o700)
}

// ProfilesPath — путь к profiles.json.
func ProfilesPath() (string, error) {
	root, err := DataRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "profiles.json"), nil
}

// LocalRepoDir — локальный клон: ...\PassDepot\repos\<profileID>.
func LocalRepoDir(profileID string) (string, error) {
	root, err := DataRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "repos", profileID), nil
}

// LocalVaultPath — файл локального профиля вне git-клонов.
func LocalVaultPath(profileID string) (string, error) {
	root, err := DataRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "vaults", profileID+".pd"), nil
}

// VaultPathInRepo — путь к файлу vault внутри git-репозитория (слэши как в git).
func VaultPathInRepo(p Profile) string {
	if p.VaultFileName != "" {
		return p.VaultFileName
	}
	return p.ID + ".pd"
}

var reVaultPrefixInvalid = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

// MakeVaultFileName делает читаемое имя файла вида "<prefix>-<id>.pd".
// Prefix строится из displayName; при невозможности — вернёт "<id>.pd".
func MakeVaultFileName(displayName, id string) string {
	p := strings.TrimSpace(displayName)
	p = strings.ReplaceAll(p, " ", "-")
	p = reVaultPrefixInvalid.ReplaceAllString(p, "-")
	p = strings.Trim(p, "-_.")
	if len(p) > 24 {
		p = p[:24]
		p = strings.Trim(p, "-_.")
	}
	if p == "" {
		return id + ".pd"
	}
	return p + "-" + id + ".pd"
}

const exportProfileNameSuffix = "-passdepot-profile.json"

// MakeExportProfileFileName — имя файла экспорта: «<имя-профиля>-passdepot-profile.json».
func MakeExportProfileFileName(displayName string) string {
	p := strings.TrimSpace(displayName)
	p = strings.ReplaceAll(p, " ", "-")
	p = reVaultPrefixInvalid.ReplaceAllString(p, "-")
	p = strings.Trim(p, "-_.")
	if len(p) > 48 {
		p = p[:48]
		p = strings.Trim(p, "-_.")
	}
	if p == "" {
		return "passdepot-profile.json"
	}
	return p + exportProfileNameSuffix
}
