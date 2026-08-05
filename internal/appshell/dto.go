package appshell

// ProfileDTO — профиль для UI (без секретов).
type ProfileDTO struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	RepoURL     string `json:"repoUrl"`
	Branch      string `json:"branch"`
	LocalOnly   bool   `json:"localOnly"`
	// HasPAT — в Credential Manager сохранён непустой GitHub PAT (нужен для входа).
	HasPAT bool `json:"hasPat"`
}

// SessionDTO — состояние после входа.
type SessionDTO struct {
	Unlocked        bool   `json:"unlocked"`
	ProfileID       string `json:"profileId"`
	DisplayName     string `json:"displayName"`
	RepoURL         string `json:"repoUrl"`
	Branch          string `json:"branch"`
	Dirty           bool   `json:"dirty"`
	EntryDirty      bool   `json:"entryDirty"`
	PendingSync     bool   `json:"pendingSync"`
	LocalOnly       bool   `json:"localOnly"`
	LastError       string `json:"lastError"`
	AutoLockMinutes int    `json:"autoLockMinutes"`
	// LastPullAt — время последнего успешного выравнивания с remote (RFC3339), пусто если ещё не было.
	LastPullAt string `json:"lastPullAt"`
}

// RemoteSaveResult — итог SaveProfileRemote.
type RemoteSaveResult struct {
	Migrated   bool   `json:"migrated"`
	OldRepoURL string `json:"oldRepoUrl"`
	NewRepoURL string `json:"newRepoUrl"`
}

// LocalVaultImportPickDTO — результат выбора .pd (ещё без создания профиля).
type LocalVaultImportPickDTO struct {
	Picked        bool   `json:"picked"`
	SuggestedName string `json:"suggestedName"`
}
