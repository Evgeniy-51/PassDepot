package profile

import (
	"os"
)

// RemoveLocalRepo удаляет каталог локального клона git (если есть). Для полного удаления профиля
// вызывайте также Remove(id) и credstore.DeletePAT(id).
func RemoveLocalRepo(profileID string) error {
	dir, err := LocalRepoDir(profileID)
	if err != nil {
		return err
	}
	return os.RemoveAll(dir)
}

// RemoveLocalVault удаляет файл локального профиля, если он существует.
func RemoveLocalVault(profileID string) error {
	path, err := LocalVaultPath(profileID)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
