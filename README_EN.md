# PassDepot

**PassDepot** is a Windows application for securely storing passwords and notes. Data is encrypted with a master password and saved in a local `.pd` file. Optionally, an encrypted copy can be synced to a private Git repository (GitHub, GitLab).

Current version: **v0.2**

---

## Table of contents

- [Features](#features)
- [System requirements](#system-requirements)
- [Installation](#installation)
- [Quick start](#quick-start)
- [Storage modes](#storage-modes)
- [Working with data](#working-with-data)
- [Import and export](#import-and-export)
- [Security](#security)
- [Where data is stored](#where-data-is-stored)
- [Building from source](#building-from-source)
- [Project architecture](#project-architecture)
- [Limitations and known quirks](#limitations-and-known-quirks)

---

## Features

- **Multiple independent profiles** — each with its own master password and vault.
- **Two storage modes:**
  - local only (no Git);
  - local + sync via a Git repository.
- **Data model:** folders → entries (title, username, password, notes).
- **Encryption** of the `.pd` vault file (Argon2id + XChaCha20-Poly1305).
- **Git sync:** save (commit + push), refresh (pull), and retry push after a network failure.
- **Import and export:**
  - encrypted `.pd` vault;
  - profile metadata as JSON (no secrets).
- **UI:** Russian and English, three appearance themes.
- **Session auto-lock** after an inactivity timeout.
- **Change master password** and rename a profile without losing data.
- **Windows installer (NSIS)** with language selection and shortcut options.

---

## System requirements

| Component | Requirement |
|-----------|-------------|
| OS | **Windows 11** (target platform) |
| WebView2 | Usually already present on Windows 11; the installer downloads the runtime if needed |
| Git | **Only for profiles with Git sync** — [Git for Windows](https://git-scm.com/download/win), available on `PATH` |
| Network | Required for clone/pull/push of Git profiles; local mode works offline |

A **local profile** does not require Git or an internet connection.

---

## Installation

### Installer (recommended)

1. Run `PassDepot-amd64-installer.exe` from `build/bin/` (or from a GitHub release).
2. Choose the installer language (Russian / English).
3. Choose the install directory (default: `C:\Program Files\PassDepot`).
4. On the final page, optionally enable:
   - Start menu shortcut;
   - desktop shortcut.
5. Finish the installation.

Uninstalling via **Apps & features** removes:

- application files;
- data under `%AppData%\PassDepot`;
- saved PATs in Windows Credential Manager.

**The remote Git repository is not modified.**

### Portable run

You can run a single `PassDepot.exe` without the installer. Data directories are created automatically on first launch (see [Where data is stored](#where-data-is-stored)).

---

## Quick start

### First launch

1. Start PassDepot.
2. Click **Add profile**.
3. On the **New profile** tab, set:
   - profile name;
   - master password (at least 8 characters) and confirmation;
   - storage mode (local or with Git).
4. Click **Save** — the vault opens.

### Creating an entry

1. Create a **folder** (New folder).
2. Select the folder.
3. Click **New entry**, fill in the fields, and confirm the changes (✓).
4. Click **Save** (or **Save file**) to write changes to the `.pd` file.

---

## Storage modes

### 1. Local vault only

- All data stays on this computer at `%AppData%\PassDepot\vaults\<id>.pd`.
- Git, a PAT, and a repository are **not required**.
- Cross-device sync is **not automatic** — move data via `.pd` export/import.

**Best for:** a single PC, offline use, minimal setup.

### 2. Local + Git repository

- Local copy: `%AppData%\PassDepot\repos\<id>\` (repository clone).
- The repository contains **only the encrypted** `.pd` file.
- Git access requires a **PAT** (Personal Access Token), stored in Windows Credential Manager.
- **Save** writes the vault locally and pushes changes to the repository (commit + push).
- **Refresh** pulls changes from the remote (only if there are no unsaved local edits).

**Best for:** multiple computers, an off-disk backup, and Git commit history.

---

## Working with data

### Main UI actions

| Action | Description |
|--------|-------------|
| **Save** | Writes changes to the local `.pd`; for a Git profile, also commit and push |
| **Refresh** | Pulls the vault from the repository (only if there are no unsaved local changes) |
| **Retry push** | Retries Git upload if the previous push failed but the file is already saved locally |
| **Log out** | Ends the session and clears decrypted data from memory |

### Sync states (Git profile)

- **Unsaved changes** — in-memory edits that have not been written to the file.
- **Not pushed to repository** — the file is saved locally, but push failed; **Retry push** is available.

### Account / profile settings

- change master password;
- rename profile;
- configure auto-lock (minutes of inactivity);
- change repository URL, branch, or PAT (with vault migration when the repository changes);
- export vault and profile;
- delete profile **locally only** (the remote repository is left untouched).

---

## Import and export

### Import

On the **Add profile** screen → **Import** tab:

| Type | What it does |
|------|----------------|
| **Import vault** (`.pd`) | Creates a new local profile from an encrypted file; you can set the name before importing |
| **Import profile** (JSON) | Adds a profile with repository settings; PAT and master password are entered separately |
| **Import from text** | Same as JSON import, via pasted content |

### Export

In the profile **Account** screen:

| Type | Contents |
|------|----------|
| **Export vault** (`.pd`) | Encrypted vault; for transfer or backup |
| **Export profile** (JSON) | Name, repository URL, branch — **no PAT or master password** |

### Moving to another computer

**Local profile:**

1. Export `.pd` on the old PC.
2. On the new PC: import `.pd`, unlock with the same master password.

**Git profile:**

1. Export the profile JSON (or create a profile with the same URL/branch).
2. On the new PC: import the JSON, enter a **new PAT** and the master password.
3. On unlock, the vault is pulled from the repository.

> The PAT is never included in exports — use a separate token with least privilege on each machine.

---

## Security

### Encryption

- Container format: binary `.pd` file (magic `PDVT`, version 1).
- **KDF:** Argon2id — derives the encryption key from the master password.
- **AEAD:** XChaCha20-Poly1305 — authenticated encryption; file tampering is detected on decrypt.
- Each save uses a fresh salt and nonce.

### Secrets

| Secret | Where it is stored |
|--------|--------------------|
| Master password | **Never persisted** — entered on every unlock |
| Vault contents | Only in the encrypted `.pd` on disk and (optionally) in Git |
| Git PAT | Windows Credential Manager (`PassDepot/PAT/<profileId>`) |
| Profile export (JSON) | No PAT or master password |

### Session

- Decrypted data exists **only in RAM** while a profile is unlocked.
- On log out, app close, or auto-lock, the session is cleared and the key and contents are wiped.
- The clipboard is cleared when you leave a profile.

### Recommendations

- Use a **long, unique** master password; without it, the `.pd` cannot be recovered.
- Store data whose disclosure could cause real harm in a **separate profile** (local vault only, with no Git repository sync).
- Do not store the master password or PAT alongside a profile export.

---

## Where data is stored

Default root directory:

```
%AppData%\PassDepot\
```

| Path | Purpose |
|------|---------|
| `profiles.json` | Profile list (names, repository URL, branch; no secrets) |
| `vaults\<profileId>.pd` | Local vault files (“local only” mode) |
| `repos\<profileId>\` | Local Git repository clones |

For testing and debugging, override the root with the `PASSDEPOT_DATA_ROOT` environment variable.

With the **NSIS installer**, `%AppData%\PassDepot` is created automatically on first use; no separate setup is required.

---

## Building from source

### Development requirements

- [Go](https://go.dev/) 1.23+
- [Node.js](https://nodejs.org/) (LTS) and npm
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) v2.12+
- For the installer: [NSIS](https://nsis.sourceforge.io/) 3.x (`makensis` on `PATH`)

### Clone and dependencies

```powershell
git clone <repository-url> PassDepot
cd PassDepot
cd frontend
npm install
cd ..
```

### Development mode

```powershell
wails dev
```

### Build the application

```powershell
wails build
```

Output: `build\bin\PassDepot.exe`

### Build the installer (NSIS)

```powershell
wails build -clean -nsis -webview2 download
```

Output: `build\bin\PassDepot-amd64-installer.exe`

> `build\windows\installer\project.nsi` must be saved as **UTF-8 with BOM**, otherwise Cyrillic text in the installer will display incorrectly.

### Tests

```powershell
go test ./...
```

---

## Project architecture

```
PassDepot/
├── main.go                 # Wails entry point; --uninstall-cleanup for the uninstaller
├── wails.json              # Wails config (name, version, frontend build)
├── frontend/               # React + TypeScript (Vite)
│   └── src/
│       ├── App.tsx         # Main UI
│       ├── i18n/           # ru/en localization
│       └── theme.ts        # Appearance themes
├── internal/
│   ├── vaultcore/          # Vault model, .pd encryption
│   ├── profile/            # Profiles, profiles.json, JSON import/export
│   ├── credstore/          # PAT in Windows Credential Manager
│   ├── gitremote/          # Wrapper around git.exe (clone, pull, push)
│   └── appshell/           # Wails API, session, sync
└── build/
    ├── bin/                # Built exe and installer
    └── windows/installer/  # NSIS installer script
```

**Stack:** Go (backend) + Wails v2 + React + WebView2.

**Git sync model:** no in-memory merge — **Refresh** fully replaces the local vault with the repository version; conflicts are resolved by “save local changes first.”

---

## Limitations and known quirks

- Platform: **Windows only** (build and testing target Windows 11).
- Git profiles require `git.exe` on PATH.
- Concurrent use of the same vault from multiple devices without coordination can diverge; the last successful push “wins” on refresh.
- Deleting a profile in the app **does not delete** the vault file from the remote repository (Git history is preserved).
- Profile names must be **unique** (case-insensitive).

---

## License

This project is distributed under the [MIT](LICENSE) license.

Copyright © 2026 Evgenii Petrashchuk
