package vaultcore

// Vault — содержимое хранилища (сериализуется в JSON внутри .pd).
type Vault struct {
	Version       int           `json:"version"` // схема JSON, не путать с версией бинарного контейнера
	Folders       []Folder      `json:"folders"`
	Descriptions  []Description `json:"descriptions"`
}

// Folder — папка с паролями.
type Folder struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
}

// Description — элемент в папке: название, логин, пароль, заметки.
// Password опционален в JSON (старые записи без поля → "").
type Description struct {
	ID        string `json:"id"`
	FolderID  string `json:"folderId"`
	Title     string `json:"title"`
	Key       string `json:"key"`      // логин
	Password  string `json:"password"` // пароль (пусто у старых записей)
	Value     string `json:"value"`    // заметки
	UpdatedAt int64  `json:"updatedAt"` // Unix seconds (Wails/JSON без time.Time)
}
