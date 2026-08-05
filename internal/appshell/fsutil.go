package appshell

import (
	"os"
	"path/filepath"

	"passdepot/internal/profile"
)

func vaultRepoPath(profileID string) (string, error) {
	repoDir, err := profile.LocalRepoDir(profileID)
	if err != nil {
		return "", err
	}
	p, ok := profile.Get(profileID)
	if ok {
		if p.VaultFileName == "" {
			p.VaultFileName = profile.MakeVaultFileName(p.DisplayName, p.ID)
		}
		return filepath.Join(repoDir, filepath.FromSlash(profile.VaultPathInRepo(p))), nil
	}
	return filepath.Join(repoDir, profileID+".pd"), nil
}

func vaultLocalPath(profileID string) (string, error) {
	p, ok := profile.Get(profileID)
	if ok && p.LocalOnly {
		return profile.LocalVaultPath(profileID)
	}
	return vaultRepoPath(profileID)
}

func atomicWriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".pd-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, path)
}

func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
