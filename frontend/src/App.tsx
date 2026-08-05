import { useCallback, useEffect, useMemo, useRef, useState, type KeyboardEvent as ReactKeyboardEvent, type MouseEvent as ReactMouseEvent } from 'react';
import '@fontsource/keania-one';
import './App.css';
import {
    AddDescription,
    AddFolder,
    ChangeMasterPassword,
    CopyToClipboard,
    CreateProfile,
    DeleteDescription,
    DeleteFolder,
    DeleteProfile,
    RenameFolder,
    RenameProfile,
    ExportLocalVaultToFile,
    ExportProfileToFile,
    GetSession,
    GetVault,
    PickLocalVaultImport,
    ConfirmLocalVaultImport,
    CancelLocalVaultImport,
    ImportProfileFromFile,
    ImportProfileJSON,
    ListProfiles,
    Login,
    Logout,
    Refresh,
    RetryPush,
    Save,
    SetAutoLockMinutes,
    SetUILocale,
    UpdateDescriptionKey,
    UpdateDescriptionPassword,
    UpdateDescriptionTitle,
    UpdateDescriptionValue,
    UpdateProfilePAT,
    SaveProfileRemote,
} from '../wailsjs/go/appshell/App';
import { BrowserOpenURL, EventsOn } from '../wailsjs/runtime/runtime';
import { appshell } from '../wailsjs/go/models';
import { vaultcore } from '../wailsjs/go/models';
import { DescDetailPanel } from './DescDetailPanel';
import { CreateProfileHelp } from './CreateProfileHelp';
import { PasswordInput } from './PasswordInput';
import { cycleTheme, readStoredTheme, THEME_LABEL, type AppTheme } from './theme';
import {
    cycleLocale,
    LOCALE_LABEL,
    readStoredLocale,
    type AppLocale,
} from './locale';
import { t, type MsgKey } from './i18n/messages';
import { FlagGb, FlagRu } from './flags';
import { APP_VERSION } from './version';

const NEW_DESC_VAL_EXPAND_ID = '__new__';

type View = 'auth' | 'vault';
type AuthMode = 'login' | 'add';

export default function App() {
    const [view, setView] = useState<View>('auth');
    const [authMode, setAuthMode] = useState<AuthMode>('login');
    const [appTheme, setAppTheme] = useState<AppTheme>(() => readStoredTheme());
    const [appLocale, setAppLocale] = useState<AppLocale>(() => readStoredLocale());
    const [profiles, setProfiles] = useState<appshell.ProfileDTO[]>([]);
    const [profileId, setProfileId] = useState('');
    const [masterPw, setMasterPw] = useState('');
    /** PAT для выбранного профиля на экране входа (импорт без токена) */
    const [loginPatInput, setLoginPatInput] = useState('');
    const [err, setErr] = useState('');
    const [note, setNote] = useState('');

    const [patUpdate, setPatUpdate] = useState('');
    const [remoteRepoURL, setRemoteRepoURL] = useState('');
    const [remoteBranch, setRemoteBranch] = useState('main');
    const [remoteMasterPw, setRemoteMasterPw] = useState('');
    const [remoteConfirmOpen, setRemoteConfirmOpen] = useState(false);
    const [migrateWarnOldUrl, setMigrateWarnOldUrl] = useState('');
    const [remoteBusy, setRemoteBusy] = useState(false);
    const [importJson, setImportJson] = useState('');
    const [importJsonExpanded, setImportJsonExpanded] = useState(false);
    const [importJsonOpen, setImportJsonOpen] = useState(false);

    const [newName, setNewName] = useState('');
    const [newURL, setNewURL] = useState('');
    const [newBranch, setNewBranch] = useState('main');
    const [newPAT, setNewPAT] = useState('');
    const [newLocalOnly, setNewLocalOnly] = useState(true);
    const [createHelpOpen, setCreateHelpOpen] = useState(false);
    const [addTab, setAddTab] = useState<'create' | 'import'>('create');
    const [importVaultNameOpen, setImportVaultNameOpen] = useState(false);
    const [importVaultName, setImportVaultName] = useState('');
    const [importVaultBusy, setImportVaultBusy] = useState(false);
    const importVaultNameRef = useRef<HTMLInputElement>(null);

    const [session, setSession] = useState<appshell.SessionDTO | null>(null);
    const [vault, setVault] = useState<vaultcore.Vault | null>(null);
    const [folderId, setFolderId] = useState('');

    const [newDescTitle, setNewDescTitle] = useState('');
    const [newDescKey, setNewDescKey] = useState('');
    const [newDescPassword, setNewDescPassword] = useState('');
    const [newDescVal, setNewDescVal] = useState('');
    const [selectedDescId, setSelectedDescId] = useState('');
    /** Редактирование / новая запись — только в модальном окне */
    const [descModalMode, setDescModalMode] = useState<null | 'new' | 'edit'>(null);
    /** Меню папки: ПКМ или «⋯» */
    const [folderMenu, setFolderMenu] = useState<{ folderId: string; x: number; y: number } | null>(null);
    const [folderCreateOpen, setFolderCreateOpen] = useState(false);
    const [newFolderName, setNewFolderName] = useState('');
    const newFolderInputRef = useRef<HTMLInputElement>(null);
    const [folderRenameId, setFolderRenameId] = useState<string | null>(null);
    const [folderRenameName, setFolderRenameName] = useState('');
    const folderRenameInputRef = useRef<HTMLInputElement>(null);
    /** id описания или NEW_DESC_VAL_EXPAND_ID — развёрнутое поле Value (textarea) */
    const [expandedValIds, setExpandedValIds] = useState<Set<string>>(() => new Set());
    /** Черновики полей до ✓ */
    const [titleDraft, setTitleDraft] = useState<Record<string, string>>({});
    const [keyDraft, setKeyDraft] = useState<Record<string, string>>({});
    const [passwordDraft, setPasswordDraft] = useState<Record<string, string>>({});
    const [valueDraft, setValueDraft] = useState<Record<string, string>>({});

    const [autoLockMin, setAutoLockMin] = useState(15);
    const [pwOld, setPwOld] = useState('');
    const [pwNew, setPwNew] = useState('');
    const [pwNew2, setPwNew2] = useState('');
    const [showAccountPw, setShowAccountPw] = useState(false);
    const [profileRenameOpen, setProfileRenameOpen] = useState(false);
    const [profileRenameName, setProfileRenameName] = useState('');
    const profileRenameInputRef = useRef<HTMLInputElement>(null);

    const [createPw, setCreatePw] = useState('');
    const [createPw2, setCreatePw2] = useState('');

    const [accountOpen, setAccountOpen] = useState(false);
    const [accountErr, setAccountErr] = useState('');
    const [pullInProgress, setPullInProgress] = useState(false);
    /** Вход / создание профиля — долгий Login (git + расшифровка) */
    const [authBusy, setAuthBusy] = useState(false);

    const tr = useCallback(
        (key: MsgKey, vars?: Record<string, string | number>) => t(appLocale, key, vars),
        [appLocale],
    );

    const selectedProfile = useMemo(() => profiles.find((p) => p.id === profileId) ?? null, [profiles, profileId]);
    const selectedHasPat = !!selectedProfile?.hasPat;
    const selectedLocalOnly = !!selectedProfile?.localOnly;

    const isProfileNameTaken = useCallback(
        (name: string, excludeId?: string) => {
            const key = name.trim().toLowerCase();
            if (!key) return false;
            return profiles.some(
                (p) => p.id !== excludeId && (p.displayName || '').trim().toLowerCase() === key,
            );
        },
        [profiles],
    );

    const lastPullLabel = useMemo(() => {
        const raw = session?.lastPullAt?.trim();
        if (!raw) {
            return '';
        }
        const d = new Date(raw);
        if (Number.isNaN(d.getTime())) {
            return raw;
        }
        return d.toLocaleString();
    }, [session?.lastPullAt]);

    const loadProfiles = useCallback(async () => {
        try {
            const list = await ListProfiles();
            setProfiles(list || []);
            setErr('');
        } catch (e: any) {
            setErr(String(e));
        }
    }, []);

    const refreshVault = useCallback(async () => {
        const v = await GetVault();
        setVault(new vaultcore.Vault(v as any));
        const s = await GetSession();
        const sd = new appshell.SessionDTO(s as any);
        setSession(sd);
        if (typeof sd.autoLockMinutes === 'number') {
            setAutoLockMin(sd.autoLockMinutes);
        }
        setFolderId((prev) => {
            const folders = (v as vaultcore.Vault)?.folders;
            if (!folders?.some((f) => f.id === prev)) {
                return '';
            }
            return prev;
        });
    }, []);

    useEffect(() => {
        loadProfiles();
    }, [loadProfiles]);

    useEffect(() => {
        void SetUILocale(readStoredLocale());
    }, []);

    useEffect(() => {
        if (!profileRenameOpen) return;
        const t = window.setTimeout(() => {
            profileRenameInputRef.current?.focus();
            profileRenameInputRef.current?.select();
        }, 0);
        return () => window.clearTimeout(t);
    }, [profileRenameOpen]);

    useEffect(() => {
        if (!importVaultNameOpen) return;
        const t = window.setTimeout(() => {
            importVaultNameRef.current?.focus();
            importVaultNameRef.current?.select();
        }, 0);
        return () => window.clearTimeout(t);
    }, [importVaultNameOpen]);

    useEffect(() => {
        if (!folderCreateOpen) return;
        const onPointerDown = (e: PointerEvent) => {
            const t = e.target;
            if (!(t instanceof Element)) return;
            const row = newFolderInputRef.current?.closest('.fRowCreate');
            if (row?.contains(t)) return;
            setFolderCreateOpen(false);
            setNewFolderName('');
        };
        document.addEventListener('pointerdown', onPointerDown);
        return () => document.removeEventListener('pointerdown', onPointerDown);
    }, [folderCreateOpen]);

    useEffect(() => {
        if (!createHelpOpen) return;
        const onKey = (e: KeyboardEvent) => {
            if (e.key === 'Escape') {
                e.preventDefault();
                setCreateHelpOpen(false);
            }
        };
        window.addEventListener('keydown', onKey);
        return () => window.removeEventListener('keydown', onKey);
    }, [createHelpOpen]);

    useEffect(() => {
        setExpandedValIds(new Set());
        setTitleDraft({});
        setKeyDraft({});
        setPasswordDraft({});
        setValueDraft({});
        setSelectedDescId('');
        setDescModalMode(null);
    }, [folderId]);

    const toggleExpandValue = useCallback((id: string) => {
        setExpandedValIds((prev) => {
            const next = new Set(prev);
            if (next.has(id)) {
                next.delete(id);
            } else {
                next.add(id);
            }
            return next;
        });
    }, []);

    useEffect(() => {
        if (profiles.length === 0) {
            setProfileId('');
            return;
        }
        setProfileId((prev) => {
            if (prev && profiles.some((p) => p.id === prev)) {
                return prev;
            }
            return profiles[0].id;
        });
    }, [profiles]);

    useEffect(() => {
        setLoginPatInput('');
    }, [profileId]);

    // Размер/показ окна делается в backend Startup() (StartHidden), чтобы убрать фликер.

    useEffect(() => {
        const off = EventsOn('session:locked', () => {
            setView('auth');
            setAuthMode('login');
            setVault(null);
            setSession(null);
            setFolderId('');
            setTitleDraft({});
            setKeyDraft({});
            setPasswordDraft({});
            setValueDraft({});
            setSelectedDescId('');
            setDescModalMode(null);
            setErr('');
            setNote(t(appLocale, 'sessionLocked'));
            window.setTimeout(() => setNote(''), 4000);
        });
        return () => off();
    }, [appLocale]);

    useEffect(() => {
        if (view === 'vault') {
            refreshVault().catch((e) => setErr(String(e)));
        }
    }, [view, refreshVault]);

    async function doSaveLoginPat() {
        if (!profileId?.trim()) {
            return;
        }
        const pat = loginPatInput.trim();
        if (!pat) {
            setErr(tr('errEnterPat'));
            return;
        }
        setErr('');
        try {
            await UpdateProfilePAT(profileId, pat);
            setLoginPatInput('');
            await loadProfiles();
            flashNote(tr('patSavedCanLogin'));
        } catch (e: any) {
            setErr(String(e));
        }
    }

    async function doLogin() {
        if (!profileId?.trim() || !masterPw) {
            return;
        }
        if (!selectedLocalOnly && !selectedHasPat) {
            setErr(tr('errSavePatFirst'));
            return;
        }
        setErr('');
        setAuthBusy(true);
        try {
            await Login(profileId, masterPw);
            setMasterPw('');
            setNote('');
            setView('vault');
        } catch (e: any) {
            setErr(String(e));
        } finally {
            setAuthBusy(false);
        }
    }

    function onAuthPasswordKeyDown(e: ReactKeyboardEvent<HTMLInputElement>) {
        if (e.key !== 'Enter') {
            return;
        }
        e.preventDefault();
        if (authBusy) {
            return;
        }
        if (profileId && masterPw.trim() && (selectedLocalOnly || selectedHasPat)) {
            void doLogin();
        }
    }

    function leaveAddMode() {
        setCreateHelpOpen(false);
        if (importVaultNameOpen) {
            void CancelLocalVaultImport();
        }
        setImportVaultNameOpen(false);
        setImportVaultName('');
        setImportVaultBusy(false);
        setNewName('');
        setCreatePw('');
        setCreatePw2('');
        setAddTab('create');
        setAuthMode('login');
    }

    function openAddMode() {
        setErr('');
        setNewName('');
        setCreatePw('');
        setCreatePw2('');
        setImportVaultNameOpen(false);
        setImportVaultName('');
        setAddTab('create');
        setAuthMode('add');
    }

    async function doCreate() {
        setErr('');
        if (createPw !== createPw2) {
            setErr(tr('errPasswordMismatch'));
            return;
        }
        if (createPw.trim().length < 8) {
            setErr(tr('errMasterMin8'));
            return;
        }
        if (isProfileNameTaken(newName)) {
            setErr(tr('errProfileNameTaken'));
            return;
        }
        const pw = createPw;
        setAuthBusy(true);
        try {
            const p = await CreateProfile(newName, newURL, newBranch, newPAT, newLocalOnly);
            setNewName('');
            setNewURL('');
            setNewPAT('');
            setNewLocalOnly(true);
            setCreatePw('');
            setCreatePw2('');
            setAddTab('create');
            setProfileId(p.id);
            await loadProfiles();
            await Login(p.id, pw);
            setNote('');
            setView('vault');
        } catch (e: any) {
            setErr(String(e));
        } finally {
            setAuthBusy(false);
        }
    }

    function clearSessionAndAuthView() {
        setErr('');
        setAccountErr('');
        Logout();
        setView('auth');
        setAuthMode('login');
        setAccountOpen(false);
        setVault(null);
        setSession(null);
        setFolderId('');
            setSelectedDescId('');
            setNewDescTitle('');
            setNewDescKey('');
            setNewDescPassword('');
            setNewDescVal('');
            setTitleDraft({});
            setKeyDraft({});
            setPasswordDraft({});
            setValueDraft({});
            setExpandedValIds(new Set());
            setDescModalMode(null);
    }

    async function doSave() {
        setErr('');
        try {
            await Save();
            await refreshVault();
        } catch (e: any) {
            setErr(String(e));
            await refreshVault();
        }
    }

    async function doRefresh() {
        setErr('');
        setPullInProgress(true);
        try {
            await Refresh();
            await refreshVault();
            flashNote(tr('vaultPulled'));
        } catch {
            await refreshVault();
        } finally {
            setPullInProgress(false);
        }
    }

    async function doRetry() {
        setErr('');
        try {
            await RetryPush();
            await refreshVault();
        } catch (e: any) {
            setErr(String(e));
            await refreshVault();
        }
    }

    async function doAddFolder() {
        if (!newFolderName.trim()) {
            return;
        }
        setErr('');
        try {
            await AddFolder(newFolderName.trim());
            await refreshVault();
            const v = await GetVault();
            const vv = new vaultcore.Vault(v as any);
            const last = vv.folders?.[vv.folders.length - 1];
            if (last?.id) setFolderId(last.id);
            setNewDescTitle('');
            setNewDescKey('');
            setNewDescPassword('');
            setNewDescVal('');
            setFolderCreateOpen(false);
            setNewFolderName('');
        } catch (e: any) {
            setErr(String(e));
        }
    }

    function openFolderCreate() {
        cancelFolderRename();
        clearFolderSelection();
        setFolderCreateOpen(true);
        setNewFolderName('');
        queueMicrotask(() => newFolderInputRef.current?.focus());
    }

    function cancelFolderCreate() {
        setFolderCreateOpen(false);
        setNewFolderName('');
    }

    function startFolderRename(id: string) {
        const f = vault?.folders?.find((x) => x.id === id);
        if (!f) {
            return;
        }
        cancelFolderCreate();
        setFolderRenameId(id);
        setFolderRenameName(f.name);
        queueMicrotask(() => {
            folderRenameInputRef.current?.focus();
            folderRenameInputRef.current?.select();
        });
    }

    function cancelFolderRename() {
        setFolderRenameId(null);
        setFolderRenameName('');
    }

    async function doAddDescription() {
        if (!folderId) {
            setErr(tr('errSelectFolder'));
            return;
        }
        setErr('');
        try {
            await AddDescription(folderId, newDescTitle, newDescKey, newDescPassword, newDescVal);
            setNewDescTitle('');
            setNewDescKey('');
            setNewDescPassword('');
            setNewDescVal('');
            setExpandedValIds((p) => {
                const n = new Set(p);
                n.delete(NEW_DESC_VAL_EXPAND_ID);
                return n;
            });
            await refreshVault();
            const v = await GetVault();
            const vv = new vaultcore.Vault(v as any);
            const list = vv.descriptions?.filter((d) => d.folderId === folderId) ?? [];
            const last = list[list.length - 1];
            if (last?.id) {
                setSelectedDescId(last.id);
            }
            setDescModalMode(null);
        } catch (e: any) {
            setErr(String(e));
        }
    }

    async function doDeleteDescription(id: string) {
        if (!confirm(tr('confirmDeleteRecord'))) return;
        setErr('');
        try {
            await DeleteDescription(id);
            setTitleDraft((p) => {
                const n = { ...p };
                delete n[id];
                return n;
            });
            setKeyDraft((p) => {
                const n = { ...p };
                delete n[id];
                return n;
            });
            setPasswordDraft((p) => {
                const n = { ...p };
                delete n[id];
                return n;
            });
            setValueDraft((p) => {
                const n = { ...p };
                delete n[id];
                return n;
            });
            if (selectedDescId === id) {
                setSelectedDescId('');
                setDescModalMode(null);
            }
            await refreshVault();
        } catch (e: any) {
            setErr(String(e));
        }
    }

    function closeFolderMenu() {
        setFolderMenu(null);
    }

    function clearFolderSelection() {
        closeFolderMenu();
        setFolderId('');
        setDescModalMode(null);
    }

    function clearFolderSelectionIfBlank(e: ReactMouseEvent) {
        const el = e.target as HTMLElement | null;
        if (!el) return;
        if (el.closest('button, input, textarea, select, a, [role="menu"], [role="menuitem"], .modal, .sheet')) {
            return;
        }
        if (!folderId) return;
        clearFolderSelection();
    }

    function openFolderMenuAt(folderId: string, anchorX: number, anchorY: number) {
        const menuW = 188;
        const menuH = 76;
        const margin = 8;
        const x = Math.max(margin, Math.min(anchorX, window.innerWidth - menuW - margin));
        const y = Math.max(margin, Math.min(anchorY, window.innerHeight - menuH - margin));
        setFolderMenu({ folderId, x, y });
    }

    async function doRenameFolder() {
        if (!folderRenameId) {
            return;
        }
        if (!folderRenameName.trim()) {
            setErr(tr('errFolderNameEmpty'));
            return;
        }
        setErr('');
        try {
            await RenameFolder(folderRenameId, folderRenameName.trim());
            cancelFolderRename();
            await refreshVault();
        } catch (e: any) {
            setErr(String(e));
        }
    }

    async function doDeleteFolder(id: string) {
        if (!confirm(tr('confirmDeleteFolder'))) return;
        setErr('');
        try {
            await DeleteFolder(id);
            closeFolderMenu();
            if (folderRenameId === id) {
                cancelFolderRename();
            }
            await refreshVault();
            const v = await GetVault();
            const vv = new vaultcore.Vault(v as any);
            setFolderId(vv.folders?.[0]?.id ?? '');
        } catch (e: any) {
            setErr(String(e));
        }
    }

    function flashNote(text: string) {
        setNote(text);
        setErr('');
        window.setTimeout(() => setNote(''), 3500);
    }

    function isRemoteTargetChanged(): boolean {
        if (!session) return false;
        return (
            remoteRepoURL.trim().replace(/\/+$/, '').toLowerCase() !==
                (session.repoUrl || '').trim().replace(/\/+$/, '').toLowerCase() ||
            (remoteBranch.trim() || 'main').toLowerCase() !== (session.branch || 'main').trim().toLowerCase()
        );
    }

    function requestSaveRemote() {
        if (!session || session.localOnly) return;
        if (!remoteRepoURL.trim() || !patUpdate.trim()) return;
        setAccountErr('');
        if (isRemoteTargetChanged()) {
            setRemoteMasterPw('');
            setRemoteConfirmOpen(true);
            return;
        }
        void doSaveRemote('');
    }

    async function doSaveRemote(masterPassword: string) {
        if (!session || session.localOnly) return;
        if (!remoteRepoURL.trim() || !patUpdate.trim()) return;
        setAccountErr('');
        setRemoteBusy(true);
        try {
            const r = await SaveProfileRemote(
                remoteRepoURL.trim(),
                remoteBranch.trim() || 'main',
                patUpdate.trim(),
                masterPassword,
            );
            setPatUpdate('');
            setRemoteMasterPw('');
            setRemoteConfirmOpen(false);
            await refreshVault();
            if (r?.migrated) {
                setMigrateWarnOldUrl(r.oldRepoUrl || '');
                flashNote(tr('repoSwitched'));
            } else {
                setMigrateWarnOldUrl('');
                flashNote(tr('patSaved'));
            }
        } catch (e: any) {
            setAccountErr(String(e));
        } finally {
            setRemoteBusy(false);
        }
    }

    async function doExportFile() {
        if (!profileId) return;
        setAccountErr('');
        try {
            await ExportProfileToFile(profileId);
            flashNote(tr('fileSaved'));
        } catch (e: any) {
            setAccountErr(String(e));
        }
    }

    async function doExportLocalVault() {
        if (!profileId) return;
        setAccountErr('');
        try {
            await ExportLocalVaultToFile(profileId);
            await refreshVault();
            flashNote(tr('vaultExported'));
        } catch (e: any) {
            setAccountErr(String(e));
        }
    }

    async function doImportFile() {
        setErr('');
        try {
            const r = await ImportProfileFromFile();
            if (!r?.id) return;
            setNewName('');
            setCreatePw('');
            setCreatePw2('');
            setAddTab('create');
            flashNote(tr('profileImported'));
            await loadProfiles();
            setProfileId(r.id);
            setAuthMode('login');
        } catch (e: any) {
            setErr(String(e));
        }
    }

    async function doImportLocalVault() {
        setErr('');
        try {
            const pick = await PickLocalVaultImport();
            if (!pick?.picked) return;
            setImportVaultName(pick.suggestedName || '');
            setImportVaultNameOpen(true);
        } catch (e: any) {
            setErr(String(e));
        }
    }

    async function cancelImportVaultName() {
        if (importVaultBusy) return;
        setImportVaultNameOpen(false);
        setImportVaultName('');
        setErr('');
        try {
            await CancelLocalVaultImport();
        } catch {
            /* ignore */
        }
    }

    async function confirmImportVaultName() {
        const name = importVaultName.trim();
        if (!name || importVaultBusy) return;
        setErr('');
        if (isProfileNameTaken(name)) {
            setErr(tr('errProfileNameTaken'));
            return;
        }
        setImportVaultBusy(true);
        try {
            const r = await ConfirmLocalVaultImport(name);
            setImportVaultNameOpen(false);
            setImportVaultName('');
            setNewName('');
            setCreatePw('');
            setCreatePw2('');
            setAddTab('create');
            flashNote(tr('localVaultImported'));
            await loadProfiles();
            setProfileId(r.id);
            setAuthMode('login');
        } catch (e: any) {
            setErr(String(e));
        } finally {
            setImportVaultBusy(false);
        }
    }

    async function doImportText() {
        if (!importJson.trim()) return;
        setErr('');
        try {
            await ImportProfileJSON(importJson.trim());
            setImportJson('');
            setNewName('');
            setCreatePw('');
            setCreatePw2('');
            setAddTab('create');
            flashNote(tr('profileImported'));
            await loadProfiles();
            setAuthMode('login');
        } catch (e: any) {
            setErr(String(e));
        }
    }

    async function copyEntryValue(val: string) {
        setErr('');
        try {
            await CopyToClipboard(val);
            flashNote(tr('copied'));
        } catch (e: any) {
            setErr(String(e));
        }
    }

    async function applyAutoLock() {
        setAccountErr('');
        try {
            let n = Math.floor(Number(autoLockMin));
            if (Number.isNaN(n) || n < 0) n = 0;
            if (n > 240) n = 240;
            await SetAutoLockMinutes(n);
            setAutoLockMin(n);
            await refreshVault();
            flashNote(tr('autoLockSaved'));
        } catch (e: any) {
            setAccountErr(String(e));
        }
    }

    async function doChangeMasterPw() {
        if (pwNew !== pwNew2) {
            setAccountErr(tr('errNewPasswordMismatch'));
            return;
        }
        if (pwNew.trim().length < 8) {
            setAccountErr(tr('errNewPasswordMin8'));
            return;
        }
        setAccountErr('');
        try {
            await ChangeMasterPassword(pwOld, pwNew);
            setPwOld('');
            setPwNew('');
            setPwNew2('');
            await refreshVault();
            flashNote(tr('masterPasswordChanged'));
        } catch (e: any) {
            setAccountErr(String(e));
        }
    }

    function openProfileRename() {
        const name = session?.displayName || selectedProfile?.displayName || '';
        setProfileRenameName(name);
        setProfileRenameOpen(true);
        setAccountErr('');
    }

    function cancelProfileRename() {
        setProfileRenameOpen(false);
        setProfileRenameName('');
        setAccountErr('');
    }

    async function doRenameProfile() {
        const id = session?.profileId || profileId;
        const name = profileRenameName.trim();
        if (!id || !name) {
            setAccountErr(tr('errEnterProfileName'));
            return;
        }
        if (isProfileNameTaken(name, id)) {
            setAccountErr(tr('errProfileNameTaken'));
            return;
        }
        setAccountErr('');
        try {
            await RenameProfile(id, name);
            setProfileRenameOpen(false);
            setProfileRenameName('');
            await loadProfiles();
            await refreshVault();
            flashNote(tr('profileRenamed'));
        } catch (e: any) {
            setAccountErr(String(e));
        }
    }

    async function doDeleteProfile() {
        const id = session?.profileId;
        if (!id) return;
        const message = session?.localOnly
            ? tr('confirmDeleteLocalProfile')
            : tr('confirmDeleteGitProfile');
        if (!confirm(message)) return;
        setAccountErr('');
        try {
            await DeleteProfile(id);
            clearSessionAndAuthView();
            setProfileId('');
            await loadProfiles();
        } catch (e: any) {
            setAccountErr(String(e));
        }
    }

    const descriptions = useMemo(
        () => vault?.descriptions?.filter((d) => d.folderId === folderId) ?? [],
        [vault, folderId],
    );

    const descriptionsSorted = useMemo(() => {
        const arr = [...descriptions];
        arr.sort((a, b) => {
            const c = a.title.localeCompare(b.title, appLocale === 'en' ? 'en' : 'ru', {
                sensitivity: 'base',
                numeric: true,
            });
            if (c !== 0) {
                return c;
            }
            return a.id.localeCompare(b.id);
        });
        return arr;
    }, [descriptions, appLocale]);

    useEffect(() => {
        setSelectedDescId((prev) => (prev && descriptions.some((d) => d.id === prev) ? prev : ''));
    }, [descriptions]);

    const selectedDesc = useMemo(
        () => descriptions.find((d) => d.id === selectedDescId) ?? null,
        [descriptions, selectedDescId],
    );

    function descriptionHasDraft(d: vaultcore.Description) {
        const t = titleDraft[d.id] ?? d.title;
        const k = keyDraft[d.id] ?? d.key;
        const p = passwordDraft[d.id] ?? d.password ?? '';
        const v = valueDraft[d.id] ?? d.value;
        return t !== d.title || k !== d.key || p !== (d.password ?? '') || v !== d.value;
    }

    function newFormDirty() {
        return !!(newDescTitle.trim() || newDescKey.trim() || newDescPassword.trim() || newDescVal.trim());
    }

    function tryCloseDescModal() {
        if (!descModalMode) {
            return;
        }
        if (descModalMode === 'new') {
            if (newFormDirty()) {
                if (!confirm(tr('confirmCloseNewDiscard'))) {
                    return;
                }
            }
            setNewDescTitle('');
            setNewDescKey('');
            setNewDescPassword('');
            setNewDescVal('');
            setExpandedValIds((p) => {
                const n = new Set(p);
                n.delete(NEW_DESC_VAL_EXPAND_ID);
                return n;
            });
            setDescModalMode(null);
            return;
        }
        if (descModalMode === 'edit' && selectedDesc && descriptionHasDraft(selectedDesc)) {
            if (!confirm(tr('confirmCloseEditDiscard'))) {
                return;
            }
        }
        setDescModalMode(null);
    }

    const tryCloseDescModalRef = useRef(tryCloseDescModal);
    tryCloseDescModalRef.current = tryCloseDescModal;

    function openDescNewModal() {
        if (descModalMode === 'edit' && selectedDesc && descriptionHasDraft(selectedDesc)) {
            if (!confirm(tr('confirmOpenNewLoseEdit'))) {
                return;
            }
        }
        if (descModalMode === 'new' && newFormDirty()) {
            if (!confirm(tr('confirmRestartNew'))) {
                return;
            }
        }
        setNewDescTitle('');
        setNewDescKey('');
        setNewDescPassword('');
        setNewDescVal('');
        setExpandedValIds((p) => {
            const n = new Set(p);
            n.delete(NEW_DESC_VAL_EXPAND_ID);
            return n;
        });
        setDescModalMode('new');
    }

    function openDescEditModal(d: vaultcore.Description) {
        if (descModalMode === 'new' && newFormDirty()) {
            if (!confirm(tr('confirmOpenDescLoseNew'))) {
                return;
            }
        }
        if (
            descModalMode === 'edit' &&
            selectedDesc &&
            selectedDesc.id !== d.id &&
            descriptionHasDraft(selectedDesc)
        ) {
            if (!confirm(tr('confirmSwitchDesc'))) {
                return;
            }
        }
        setSelectedDescId(d.id);
        setDescModalMode('edit');
    }

    useEffect(() => {
        if (!descModalMode) {
            return undefined;
        }
        const onKey = (e: globalThis.KeyboardEvent) => {
            if (e.key === 'Escape') {
                e.preventDefault();
                tryCloseDescModalRef.current();
            }
        };
        window.addEventListener('keydown', onKey);
        return () => window.removeEventListener('keydown', onKey);
    }, [descModalMode]);

    useEffect(() => {
        if (descModalMode === 'edit' && selectedDescId && !descriptions.some((d) => d.id === selectedDescId)) {
            setDescModalMode(null);
        }
    }, [descModalMode, selectedDescId, descriptions]);

    useEffect(() => {
        if (!folderMenu) {
            return undefined;
        }
        const onKey = (e: globalThis.KeyboardEvent) => {
            if (e.key === 'Escape') {
                e.preventDefault();
                setFolderMenu(null);
            }
        };
        window.addEventListener('keydown', onKey);
        return () => window.removeEventListener('keydown', onKey);
    }, [folderMenu]);

    const hasLocalDescDraft = useMemo(
        () =>
            descriptions.some((d) => {
                const t = titleDraft[d.id] ?? d.title;
                const k = keyDraft[d.id] ?? d.key;
                const p = passwordDraft[d.id] ?? d.password ?? '';
                const v = valueDraft[d.id] ?? d.value;
                return t !== d.title || k !== d.key || p !== (d.password ?? '') || v !== d.value;
            }),
        [descriptions, titleDraft, keyDraft, passwordDraft, valueDraft],
    );

    async function requestLogout() {
        const backendUnsaved = !!(session?.dirty || session?.entryDirty);
        const draftsUnsaved = hasLocalDescDraft;

        if (!backendUnsaved && !draftsUnsaved) {
            clearSessionAndAuthView();
            return;
        }

        if (!backendUnsaved && draftsUnsaved) {
            if (!confirm(tr('confirmLogoutDrafts'))) {
                return;
            }
            clearSessionAndAuthView();
            return;
        }

        if (confirm(session?.localOnly ? tr('confirmLogoutSaveLocal') : tr('confirmLogoutSaveGit'))) {
            setErr('');
            try {
                await Save();
                await refreshVault();
            } catch (e: any) {
                setErr(String(e));
                await refreshVault();
                return;
            }
            const v = await GetVault();
            const vv = new vaultcore.Vault(v as any);
            const ents = vv.descriptions?.filter((e) => e.folderId === folderId) ?? [];
            const stillDraft = ents.some((e) => {
                const t = titleDraft[e.id] ?? e.title;
                const k = keyDraft[e.id] ?? e.key;
                const p = passwordDraft[e.id] ?? e.password ?? '';
                const v = valueDraft[e.id] ?? e.value;
                return t !== e.title || k !== e.key || p !== (e.password ?? '') || v !== e.value;
            });
            if (stillDraft) {
                if (!confirm(tr('confirmLogoutAfterSaveDrafts'))) {
                    return;
                }
            }
            clearSessionAndAuthView();
            return;
        }

        if (
            !confirm(
                tr('confirmLogoutDiscardFile') + (draftsUnsaved ? tr('confirmLogoutDiscardFileDrafts') : ''),
            )
        ) {
            return;
        }
        clearSessionAndAuthView();
    }

    function discardTitleDraft(id: string) {
        setTitleDraft((p) => {
            const n = { ...p };
            delete n[id];
            return n;
        });
    }
    function discardKeyDraft(id: string) {
        setKeyDraft((p) => {
            const n = { ...p };
            delete n[id];
            return n;
        });
    }
    function discardPasswordDraft(id: string) {
        setPasswordDraft((p) => {
            const n = { ...p };
            delete n[id];
            return n;
        });
    }
    function discardValueDraft(id: string) {
        setValueDraft((p) => {
            const n = { ...p };
            delete n[id];
            return n;
        });
    }

    async function applyTitleDraft(descId: string) {
        const d = descriptions.find((x) => x.id === descId);
        if (!d) return;
        const val = titleDraft[descId] ?? d.title;
        if (!val.trim()) {
            setErr(tr('errTitleEmpty'));
            return;
        }
        setErr('');
        try {
            await UpdateDescriptionTitle(descId, val);
            setTitleDraft((p) => {
                const n = { ...p };
                delete n[descId];
                return n;
            });
            await refreshVault();
            flashNote(session?.localOnly ? tr('changeSavedLocal') : tr('changeSavedGit'));
        } catch (e2: any) {
            setErr(String(e2));
        }
    }

    async function applyKeyDraft(descId: string) {
        const d = descriptions.find((x) => x.id === descId);
        if (!d) return;
        const val = keyDraft[descId] ?? d.key;
        setErr('');
        try {
            await UpdateDescriptionKey(descId, val);
            setKeyDraft((p) => {
                const n = { ...p };
                delete n[descId];
                return n;
            });
            await refreshVault();
            flashNote(session?.localOnly ? tr('changeSavedLocal') : tr('changeSavedGit'));
        } catch (e2: any) {
            setErr(String(e2));
        }
    }

    async function applyPasswordDraft(descId: string) {
        const d = descriptions.find((x) => x.id === descId);
        if (!d) return;
        const val = passwordDraft[descId] ?? d.password ?? '';
        setErr('');
        try {
            await UpdateDescriptionPassword(descId, val);
            setPasswordDraft((p) => {
                const n = { ...p };
                delete n[descId];
                return n;
            });
            await refreshVault();
            flashNote(session?.localOnly ? tr('changeSavedLocal') : tr('changeSavedGit'));
        } catch (e2: any) {
            setErr(String(e2));
        }
    }

    async function applyValueDraft(descId: string) {
        const d = descriptions.find((x) => x.id === descId);
        if (!d) return;
        const val = valueDraft[descId] ?? d.value;
        setErr('');
        try {
            await UpdateDescriptionValue(descId, val);
            setValueDraft((p) => {
                const n = { ...p };
                delete n[descId];
                return n;
            });
            await refreshVault();
            flashNote(session?.localOnly ? tr('changeSavedLocal') : tr('changeSavedGit'));
        } catch (e2: any) {
            setErr(String(e2));
        }
    }

    /** Есть что писать в .pd / отправить в git (как проверка в backend Save). */
    const canSaveFile = !!(session?.dirty || session?.pendingSync);

    const showVaultToolbarStatus =
        !!session?.entryDirty ||
        !!session?.pendingSync ||
        pullInProgress ||
        !!lastPullLabel ||
        hasLocalDescDraft;

    const vaultModalCoversBanners = view === 'vault' && (accountOpen || !!descModalMode);
    const showGlobalBanners = !vaultModalCoversBanners;

    return (
        <div id="App" className={view === 'auth' ? 'wrap auth' : 'wrap vaultApp'}>
            {view !== 'auth' && (
                <header className="hdr authHdr">
                    <h1 className="logo">PassDepot</h1>
                </header>
            )}

            {err && showGlobalBanners && (
                <div className="err errGlobal" role="alert">
                    <div className="errGlobalBody">{err}</div>
                    <button
                        type="button"
                        className="errDismiss"
                        onClick={() => setErr('')}
                        aria-label={tr('closeNotification')}
                    >
                        ×
                    </button>
                </div>
            )}
            {note && showGlobalBanners && <div className="ok">{note}</div>}

            {view === 'auth' && (
                <div className="authStack">
                    <header className="hdr authHdr">
                        <h1 className="logo">PassDepot</h1>
                        <span className="appVersion" title={tr('versionTitle')}>
                            {APP_VERSION}
                        </span>
                    </header>
                    <main className="authMain">
                    <div className={`panel authCard${authMode === 'add' ? ' authCardAdd' : ''}`}>
                        {authMode === 'login' && (
                            <>
                                <div className="authForm">
                                    {authBusy && (
                                        <div className="authBusyBanner" role="status" aria-live="polite">
                                            <span className="authSpinner" aria-hidden />
                                            {selectedLocalOnly ? tr('openingLocal') : tr('openingRemote')}
                                        </div>
                                    )}
                                    <div className="authRow authRowProfile">
                                        <div className="authRowMain">
                                            <label className="albl" htmlFor="authProfileSelect">
                                                {tr('selectProfile')}
                                            </label>
                                            <div className="selectWrap authRowSelect">
                                                <select
                                                    id="authProfileSelect"
                                                    className="ainp ainpSelect"
                                                    value={profileId}
                                                    onChange={(e) => setProfileId(e.target.value)}
                                                    disabled={profiles.length === 0 || authBusy}
                                                >
                                                    {profiles.length === 0 ? (
                                                        <option value="" disabled>
                                                            {tr('addProfileOption')}
                                                        </option>
                                                    ) : (
                                                        profiles.map((p) => (
                                                            <option key={p.id} value={p.id}>
                                                                {p.displayName}
                                                            </option>
                                                        ))
                                                    )}
                                                </select>
                                            </div>
                                        </div>
                                        <button
                                            className="abtn abtnRow pdBtn pdBtn--secondary"
                                            type="button"
                                            onClick={openAddMode}
                                            disabled={authBusy}
                                        >
                                            {tr('addProfile')}
                                        </button>
                                    </div>

                                    {profileId && !selectedLocalOnly && !selectedHasPat && (
                                        <div className="authPatMissing">
                                            <p className="authPatMissingTxt">{tr('patMissingHint')}</p>
                                            <div className="authPatRow">
                                                <input
                                                    className="ainp authPatInp"
                                                    type="text"
                                                    placeholder={tr('patPlaceholder')}
                                                    value={loginPatInput}
                                                    onChange={(e) => setLoginPatInput(e.target.value)}
                                                    autoComplete="off"
                                                    spellCheck={false}
                                                    disabled={authBusy}
                                                />
                                                <button
                                                    className="abtn abtnPat pdBtn pdBtn--primary"
                                                    type="button"
                                                    onClick={() => void doSaveLoginPat()}
                                                    disabled={authBusy || !loginPatInput.trim()}
                                                >
                                                    {tr('savePat')}
                                                </button>
                                            </div>
                                        </div>
                                    )}

                                    <div className="authRow authRowLogin">
                                        <div className="authRowMain">
                                            <label className="albl" htmlFor="authMasterPassword">
                                                {tr('password')}
                                            </label>
                                            <PasswordInput
                                                id="authMasterPassword"
                                                inputClassName="ainp ainpGrow"
                                                value={masterPw}
                                                onChange={(e) => setMasterPw(e.target.value)}
                                                onKeyDown={onAuthPasswordKeyDown}
                                                autoComplete="off"
                                                disabled={authBusy}
                                                placeholder={tr('masterPassword')}
                                                locale={appLocale}
                                            />
                                        </div>
                                        <button
                                            className="abtn abtnRow abtnLogin pdBtn pdBtn--primary"
                                            type="button"
                                            onClick={() => void doLogin()}
                                            disabled={
                                                !profileId ||
                                                !masterPw.trim() ||
                                                authBusy ||
                                                (!selectedLocalOnly && !selectedHasPat)
                                            }
                                            title={
                                                !selectedLocalOnly && !selectedHasPat
                                                    ? tr('savePatFirstTitle')
                                                    : undefined
                                            }
                                        >
                                            {authBusy ? tr('signingIn') : tr('signIn')}
                                        </button>
                                    </div>
                                </div>
                            </>
                        )}

                        {authMode === 'add' && (
                            <div className="authForm authFormAdd">
                                <div className="authAddHdr">
                                    <div className="authAddTabs" role="tablist" aria-label={tr('addProfile')}>
                                        <button
                                            className={
                                                'authAddTab' + (addTab === 'create' ? ' authAddTab--active' : '')
                                            }
                                            type="button"
                                            role="tab"
                                            aria-selected={addTab === 'create'}
                                            onClick={() => setAddTab('create')}
                                            disabled={authBusy}
                                        >
                                            {tr('addTabNew')}
                                        </button>
                                        <button
                                            className={
                                                'authAddTab' + (addTab === 'import' ? ' authAddTab--active' : '')
                                            }
                                            type="button"
                                            role="tab"
                                            aria-selected={addTab === 'import'}
                                            onClick={() => setAddTab('import')}
                                            disabled={authBusy}
                                        >
                                            {tr('addTabImport')}
                                        </button>
                                    </div>
                                    <button
                                        className="xbtn authAddClose"
                                        type="button"
                                        onClick={leaveAddMode}
                                        disabled={authBusy}
                                        aria-label={tr('close')}
                                    >
                                        ×
                                    </button>
                                </div>

                                {authBusy && addTab === 'create' && (
                                    <div className="authBusyBanner" role="status" aria-live="polite">
                                        <span className="authSpinner" aria-hidden />
                                        {newLocalOnly ? tr('creatingLocal') : tr('creatingRemote')}
                                    </div>
                                )}

                                {addTab === 'create' && (
                                    <>
                                        <label className="lbl">{tr('profileName')}</label>
                                        <input
                                            className="inp"
                                            value={newName}
                                            onChange={(e) => setNewName(e.target.value)}
                                            disabled={authBusy}
                                        />
                                        <label className="lbl lblWithHint">
                                            {tr('masterPassword')}
                                            <span className="fieldHint">{tr('masterPasswordHint')}</span>
                                        </label>
                                        <PasswordInput
                                            inputClassName="inp"
                                            value={createPw}
                                            onChange={(e) => setCreatePw(e.target.value)}
                                            autoComplete="off"
                                            disabled={authBusy}
                                            locale={appLocale}
                                        />
                                        <label className="lbl">{tr('masterPasswordRepeat')}</label>
                                        <PasswordInput
                                            inputClassName="inp"
                                            value={createPw2}
                                            onChange={(e) => setCreatePw2(e.target.value)}
                                            autoComplete="off"
                                            disabled={authBusy}
                                            locale={appLocale}
                                        />
                                        <div className="authGitBlock">
                                            <div
                                                className="storageModeRow"
                                                role="radiogroup"
                                                aria-label={tr('storageTypeAria')}
                                            >
                                                <label className="storageModeOpt">
                                                    <input
                                                        type="radio"
                                                        name="storageMode"
                                                        checked={newLocalOnly}
                                                        onChange={() => setNewLocalOnly(true)}
                                                        disabled={authBusy}
                                                    />
                                                    <span>{tr('storageLocalOnly')}</span>
                                                </label>
                                                <label className="storageModeOpt">
                                                    <input
                                                        type="radio"
                                                        name="storageMode"
                                                        checked={!newLocalOnly}
                                                        onChange={() => setNewLocalOnly(false)}
                                                        disabled={authBusy}
                                                    />
                                                    <span>{tr('storageLocalAndGit')}</span>
                                                </label>
                                            </div>
                                            <div
                                                className={
                                                    'authGitFields' +
                                                    (newLocalOnly ? ' authGitFields--inactive' : '')
                                                }
                                            >
                                                <label className="lbl">{tr('repoUrlLabel')}</label>
                                                <input
                                                    className="inp"
                                                    value={newURL}
                                                    onChange={(e) => setNewURL(e.target.value)}
                                                    disabled={authBusy || newLocalOnly}
                                                />
                                                <label className="lbl">{tr('branchLabel')}</label>
                                                <input
                                                    className="inp"
                                                    value={newBranch}
                                                    onChange={(e) => setNewBranch(e.target.value)}
                                                    disabled={authBusy || newLocalOnly}
                                                />
                                                <label className="lbl">{tr('patPlaceholder')}</label>
                                                <input
                                                    className="inp"
                                                    type="text"
                                                    value={newPAT}
                                                    onChange={(e) => setNewPAT(e.target.value)}
                                                    autoComplete="off"
                                                    spellCheck={false}
                                                    disabled={authBusy || newLocalOnly}
                                                />
                                            </div>
                                        </div>
                                        <div className="row authCreateActions">
                                            <button
                                                className="btn pdBtn pdBtn--primary"
                                                onClick={doCreate}
                                                disabled={
                                                    !newName.trim() ||
                                                    (!newLocalOnly &&
                                                        (!newURL.trim() ||
                                                            !newBranch.trim() ||
                                                            !newPAT.trim())) ||
                                                    !createPw.trim() ||
                                                    !createPw2.trim() ||
                                                    authBusy
                                                }
                                            >
                                                {authBusy ? tr('saving') : tr('save')}
                                            </button>
                                            <button
                                                className="btn pdBtn pdBtn--secondary"
                                                type="button"
                                                onClick={leaveAddMode}
                                                disabled={authBusy}
                                            >
                                                {tr('cancelVerb')}
                                            </button>
                                            <button
                                                className="btn pdBtn pdBtn--secondary authCreateHelpBtn"
                                                type="button"
                                                onClick={() => setCreateHelpOpen(true)}
                                                disabled={authBusy}
                                            >
                                                {tr('help')}
                                            </button>
                                        </div>
                                    </>
                                )}

                                {addTab === 'import' && (
                                    <div className="authAddImportPane" role="tabpanel">
                                        <div className="authImportList">
                                            <div className="authImportRow">
                                                <button
                                                    className="btn pdBtn pdBtn--secondary authImportBtn"
                                                    onClick={doImportLocalVault}
                                                    disabled={authBusy}
                                                >
                                                    {tr('importVault')}
                                                </button>
                                                <span className="authImportHint">{tr('importVaultHint')}</span>
                                            </div>
                                            <div className="authImportRow">
                                                <button
                                                    className="btn pdBtn pdBtn--secondary authImportBtn"
                                                    onClick={doImportFile}
                                                    disabled={authBusy}
                                                >
                                                    {tr('importProfile')}
                                                </button>
                                                <span className="authImportHint">{tr('importProfileHint')}</span>
                                            </div>
                                        </div>
                                        <button
                                            className="authImportMoreBtn"
                                            type="button"
                                            onClick={() => setImportJsonOpen((v) => !v)}
                                            disabled={authBusy}
                                            aria-expanded={importJsonOpen}
                                        >
                                            {importJsonOpen ? tr('importJsonLess') : tr('importJsonMore')}
                                        </button>
                                        {importJsonOpen && (
                                            <div className="authImportMore">
                                                <label className="lbl">{tr('importJsonLabel')}</label>
                                                <textarea
                                                    className="inp ta authImportTa"
                                                    data-expanded={importJsonExpanded ? '1' : '0'}
                                                    rows={importJsonExpanded ? 10 : 2}
                                                    value={importJson}
                                                    onChange={(e) => setImportJson(e.target.value)}
                                                    onDoubleClick={() => setImportJsonExpanded((v) => !v)}
                                                    onPaste={(e) => {
                                                        const text = e.clipboardData.getData('text');
                                                        if (text.includes('\n') || text.length > 120) {
                                                            setImportJsonExpanded(true);
                                                        }
                                                    }}
                                                    placeholder={tr('importJsonPlaceholder')}
                                                    title={tr('dblClickExpandTitle')}
                                                    disabled={authBusy}
                                                />
                                                <button
                                                    className="btn pdBtn pdBtn--secondary"
                                                    onClick={doImportText}
                                                    disabled={!importJson.trim() || authBusy}
                                                >
                                                    {tr('importFromText')}
                                                </button>
                                            </div>
                                        )}
                                    </div>
                                )}
                            </div>
                        )}
                    </div>
                </main>
                </div>
            )}

            <CreateProfileHelp open={createHelpOpen} onClose={() => setCreateHelpOpen(false)} locale={appLocale} />

            {importVaultNameOpen && (
                <div className="modal importVaultNameModal" role="dialog" aria-modal="true" aria-labelledby="importVaultNameTitle">
                    <div className="backdrop" onClick={() => void cancelImportVaultName()} />
                    <div className="sheet sheetImportVaultName">
                        <div className="sheetHdr">
                            <div className="sheetTitle" id="importVaultNameTitle">
                                {tr('importVaultNameTitle')}
                            </div>
                        </div>
                        <p className="importVaultNameHint">{tr('importVaultNameHint')}</p>
                        <label className="lbl" htmlFor="importVaultNameInp">
                            {tr('profileName')}
                        </label>
                        <input
                            id="importVaultNameInp"
                            ref={importVaultNameRef}
                            className="inp"
                            type="text"
                            value={importVaultName}
                            onChange={(e) => setImportVaultName(e.target.value)}
                            disabled={importVaultBusy}
                            onKeyDown={(e) => {
                                if (e.key === 'Enter') {
                                    e.preventDefault();
                                    void confirmImportVaultName();
                                } else if (e.key === 'Escape') {
                                    e.preventDefault();
                                    void cancelImportVaultName();
                                }
                            }}
                        />
                        {err && (
                            <div className="err sheetBanner" role="alert">
                                {err}
                            </div>
                        )}
                        <div className="importVaultNameActions">
                            <button
                                className="btn pdBtn pdBtn--primary"
                                type="button"
                                disabled={importVaultBusy || !importVaultName.trim()}
                                onClick={() => void confirmImportVaultName()}
                            >
                                {importVaultBusy ? tr('saving') : tr('importVaultConfirm')}
                            </button>
                            <button
                                className="btn pdBtn pdBtn--secondary"
                                type="button"
                                disabled={importVaultBusy}
                                onClick={() => void cancelImportVaultName()}
                            >
                                {tr('cancel')}
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {view === 'auth' && (
                <footer className="authFooter">
                    <button
                        type="button"
                        className="authFooterBtn authLocaleBtn"
                        title={appLocale === 'ru' ? tr('localeTitleRu') : tr('localeTitleEn')}
                        onClick={() => {
                            const next = cycleLocale();
                            setAppLocale(next);
                            void SetUILocale(next);
                        }}
                    >
                        <span className="authLocaleFlag" aria-hidden="true">
                            {appLocale === 'ru' ? <FlagRu /> : <FlagGb />}
                        </span>
                        <span className="authLocaleName">{LOCALE_LABEL[appLocale]}</span>
                    </button>
                    <button
                        type="button"
                        className="authFooterBtn"
                        title={tr('themeTitle', { theme: THEME_LABEL[appTheme] })}
                        onClick={() => {
                            const next = cycleTheme();
                            setAppTheme(next);
                        }}
                    >
                        {tr('themePrefix', { theme: THEME_LABEL[appTheme] })}
                    </button>
                </footer>
            )}

            {view === 'vault' && vault && (
                <main className="vaultMain">
                    <div className="panel vault">
                        <div className="toolbar">
                            <div className="toolbarActions">
                                <button
                                    className="pbtn"
                                    type="button"
                                    onClick={() => {
                                        setAccountErr('');
                                        setPwOld('');
                                        setPwNew('');
                                        setPwNew2('');
                                        setShowAccountPw(false);
                                        setProfileRenameOpen(false);
                                        setProfileRenameName('');
                                        setPatUpdate('');
                                        setRemoteMasterPw('');
                                        setRemoteConfirmOpen(false);
                                        setMigrateWarnOldUrl('');
                                        setRemoteRepoURL(session?.repoUrl || '');
                                        setRemoteBranch(session?.branch || 'main');
                                        setAccountOpen(true);
                                    }}
                                    title={
                                        session?.displayName
                                            ? tr('accountTitleNamed', { name: session.displayName })
                                            : tr('accountTitle')
                                    }
                                >
                                    <span className="pico" aria-hidden="true">
                                        <svg viewBox="0 0 24 24" width="18" height="18">
                                            <path
                                                d="M12 12a4.5 4.5 0 1 0-4.5-4.5A4.51 4.51 0 0 0 12 12Zm0 2c-4.42 0-8 2.24-8 5v1h16v-1c0-2.76-3.58-5-8-5Z"
                                                fill="currentColor"
                                            />
                                        </svg>
                                    </span>
                                    <span className="ptxt">{session?.displayName}</span>
                                </button>
                                <button
                                    className="btn saveVault tbEq"
                                    type="button"
                                    onClick={doSave}
                                    disabled={!canSaveFile}
                                    title={canSaveFile ? tr('saveChangesTitle') : tr('nothingToSaveTitle')}
                                >
                                    {tr('save')}
                                </button>
                                {!session?.localOnly && (
                                    <>
                                        <button
                                            className="btn tbEq"
                                            type="button"
                                            onClick={() => void doRefresh()}
                                            disabled={!!session?.dirty || pullInProgress}
                                            title={
                                                session?.dirty
                                                    ? tr('refreshBlockedDirty')
                                                    : pullInProgress
                                                      ? tr('refreshBusy')
                                                      : tr('refreshTitle')
                                            }
                                        >
                                            {tr('refresh')}
                                        </button>
                                        <button
                                            className="btn tbEq"
                                            type="button"
                                            onClick={doRetry}
                                            disabled={!session?.pendingSync}
                                            title={session?.pendingSync ? tr('retryPushTitle') : tr('retryPushNone')}
                                        >
                                            {tr('retryPush')}
                                        </button>
                                    </>
                                )}
                                <button
                                    className="btn tbLogout"
                                    type="button"
                                    onClick={() => void requestLogout()}
                                    title={tr('logoutTitle')}
                                >
                                    {tr('logout')}
                                </button>
                            </div>
                            {showVaultToolbarStatus && (
                                <div className="toolbarStatus" aria-live="polite">
                                    {session?.entryDirty && <span className="warn">{tr('warnUnsaved')}</span>}
                                    {hasLocalDescDraft && <span className="warn">{tr('warnFieldDraft')}</span>}
                                    {!session?.localOnly && session?.pendingSync && (
                                        <span className="warn">{tr('warnPendingSync')}</span>
                                    )}
                                    {!session?.localOnly && pullInProgress && (
                                        <span className="syncBusy">{tr('syncBusy')}</span>
                                    )}
                                    {!session?.localOnly && !pullInProgress && lastPullLabel && (
                                        <span className="syncMeta" title={session?.lastPullAt}>
                                            {tr('syncMeta', { when: lastPullLabel })}
                                        </span>
                                    )}
                                </div>
                            )}
                        </div>
                        {!session?.localOnly && session?.lastError && !accountOpen && !descModalMode && (
                            <div className="err soft" role="status">
                                {tr('syncErrorPrefix')}
                                {session.lastError}
                            </div>
                        )}

                        {accountOpen && (
                            <div className="modal" role="dialog" aria-modal="true">
                                <div
                                    className="backdrop"
                                    onClick={() => {
                                        setAccountErr('');
                                        setAccountOpen(false);
                                    }}
                                />
                                <div className="sheet sheetAccount">
                                    <div className="sheetHdr">
                                        <div className="sheetTitle">{tr('account')}</div>
                                        <button
                                            className="xbtn"
                                            type="button"
                                            onClick={() => {
                                                setAccountErr('');
                                                setAccountOpen(false);
                                            }}
                                            aria-label={tr('close')}
                                        >
                                            ×
                                        </button>
                                    </div>
                                    {(err || accountErr || (!session?.localOnly && session?.lastError) || note) && (
                                        <div className="sheetNotifications">
                                            {err && (
                                                <div className="err sheetBanner" role="alert">
                                                    {err}
                                                </div>
                                            )}
                                            {accountErr && (
                                                <div className="err sheetBanner" role="alert">
                                                    {accountErr}
                                                </div>
                                            )}
                                            {!session?.localOnly && session?.lastError && (
                                                <div className="err soft sheetBanner" role="status">
                                                    {tr('syncErrorPrefix')}
                                                    {session.lastError}
                                                </div>
                                            )}
                                            {note && (
                                                <div className="ok sheetBanner" role="status">
                                                    {note}
                                                </div>
                                            )}
                                        </div>
                                    )}

                                    <div className="row settings accountProfileRow">
                                        {profileRenameOpen ? (
                                            <div className="accountProfileEdit">
                                                <input
                                                    ref={profileRenameInputRef}
                                                    className="inp accountProfileRenameInp"
                                                    type="text"
                                                    placeholder={tr('newProfileNamePlaceholder')}
                                                    value={profileRenameName}
                                                    onChange={(e) => setProfileRenameName(e.target.value)}
                                                    onKeyDown={(e) => {
                                                        if (e.key === 'Enter') {
                                                            e.preventDefault();
                                                            void doRenameProfile();
                                                        } else if (e.key === 'Escape') {
                                                            e.preventDefault();
                                                            cancelProfileRename();
                                                        }
                                                    }}
                                                    autoComplete="off"
                                                />
                                                <div className="accountProfileEditActions">
                                                    <button
                                                        className="btn"
                                                        type="button"
                                                        disabled={!profileRenameName.trim()}
                                                        onClick={() => void doRenameProfile()}
                                                    >
                                                        {tr('save')}
                                                    </button>
                                                    <button
                                                        className="btn"
                                                        type="button"
                                                        onClick={cancelProfileRename}
                                                    >
                                                        {tr('cancelVerb')}
                                                    </button>
                                                </div>
                                            </div>
                                        ) : (
                                            <button
                                                type="button"
                                                className="accountProfileNameBtn"
                                                title={tr('renameProfileTitle')}
                                                onClick={openProfileRename}
                                            >
                                                {session?.displayName ||
                                                    selectedProfile?.displayName ||
                                                    session?.profileId ||
                                                    ''}
                                            </button>
                                        )}
                                    </div>

                                    <div className="row settings accountAutoLockRow">
                                        <span className="accountAutoLockTxt">{tr('autoLockBefore')}</span>
                                        <input
                                            className="inp narrow accountAutoLockInp"
                                            type="number"
                                            min={0}
                                            max={240}
                                            value={autoLockMin}
                                            onChange={(e) => setAutoLockMin(Number(e.target.value))}
                                            aria-label={tr('autoLockAria')}
                                        />
                                        <span className="accountAutoLockTxt">{tr('autoLockAfter')}</span>
                                        <button className="btn accountAutoLockBtn" type="button" onClick={applyAutoLock}>
                                            {tr('apply')}
                                        </button>
                                    </div>

                                    <div className="row settings accountPwBlock">
                                        <label className="lbl inline">{tr('changeMasterPassword')}</label>
                                        <div className="accountPwFields">
                                            <input
                                                className="inp accountPwField"
                                                type={showAccountPw ? 'text' : 'password'}
                                                placeholder={tr('currentPassword')}
                                                value={pwOld}
                                                onChange={(e) => setPwOld(e.target.value)}
                                                autoComplete="off"
                                            />
                                            <input
                                                className="inp accountPwField"
                                                type={showAccountPw ? 'text' : 'password'}
                                                placeholder={tr('newPasswordMin')}
                                                value={pwNew}
                                                onChange={(e) => setPwNew(e.target.value)}
                                                autoComplete="off"
                                            />
                                            <input
                                                className="inp accountPwField"
                                                type={showAccountPw ? 'text' : 'password'}
                                                placeholder={tr('passwordAgain')}
                                                value={pwNew2}
                                                onChange={(e) => setPwNew2(e.target.value)}
                                                autoComplete="off"
                                            />
                                            <div className="accountPwFooter">
                                                <button
                                                    className="btn accountBlockBtn"
                                                    type="button"
                                                    onClick={doChangeMasterPw}
                                                >
                                                    {tr('change')}
                                                </button>
                                                <label className="accountShowPw">
                                                    <input
                                                        type="checkbox"
                                                        checked={showAccountPw}
                                                        onChange={(e) => setShowAccountPw(e.target.checked)}
                                                    />
                                                    <span>{tr('showInput')}</span>
                                                </label>
                                            </div>
                                        </div>
                                    </div>

                                    {session && !session.localOnly && (
                                        <div className="row settings accountRemoteBlock">
                                            <label className="lbl inline">{tr('repository')}</label>
                                            {session.repoUrl && (
                                                <div className="accountRemoteCurrent">
                                                    <span className="accountRemoteCurrentLbl">{tr('currently')}</span>
                                                    <button
                                                        type="button"
                                                        className="accountRemoteLink"
                                                        title={tr('openInBrowser')}
                                                        onClick={() => BrowserOpenURL(session.repoUrl)}
                                                    >
                                                        {session.repoUrl}
                                                    </button>
                                                    {session.branch ? (
                                                        <span className="accountRemoteBranch">@{session.branch}</span>
                                                    ) : null}
                                                </div>
                                            )}
                                            <input
                                                className="inp accountRemoteField"
                                                placeholder={tr('repoUrlLabel')}
                                                value={remoteRepoURL}
                                                onChange={(e) => setRemoteRepoURL(e.target.value)}
                                                autoComplete="off"
                                                disabled={remoteBusy}
                                            />
                                            <input
                                                className="inp accountRemoteField accountRemoteBranchInp"
                                                placeholder={tr('branchLabel')}
                                                value={remoteBranch}
                                                onChange={(e) => setRemoteBranch(e.target.value)}
                                                autoComplete="off"
                                                disabled={remoteBusy}
                                            />
                                            <input
                                                className="inp accountRemoteField"
                                                type="text"
                                                placeholder={tr('patPlaceholder')}
                                                value={patUpdate}
                                                onChange={(e) => setPatUpdate(e.target.value)}
                                                autoComplete="off"
                                                spellCheck={false}
                                                disabled={remoteBusy}
                                            />
                                            <p className="accountRemoteHint">{tr('remoteHint')}</p>
                                            <button
                                                className="btn accountBlockBtn"
                                                type="button"
                                                onClick={() => requestSaveRemote()}
                                                disabled={remoteBusy || !remoteRepoURL.trim() || !patUpdate.trim()}
                                            >
                                                {remoteBusy ? tr('saving') : tr('save')}
                                            </button>
                                            {migrateWarnOldUrl && (
                                                <div className="accountMigrateWarn" role="status">
                                                    <p>{tr('migrateWarn')}</p>
                                                    <button
                                                        type="button"
                                                        className="accountRemoteLink"
                                                        title={tr('openOldRepo')}
                                                        onClick={() => BrowserOpenURL(migrateWarnOldUrl)}
                                                    >
                                                        {migrateWarnOldUrl}
                                                    </button>
                                                </div>
                                            )}
                                        </div>
                                    )}

                                    {remoteConfirmOpen && (
                                        <div className="modal accountRemoteConfirmModal" role="dialog" aria-modal="true">
                                            <div
                                                className="backdrop"
                                                onClick={() => {
                                                    if (remoteBusy) return;
                                                    setRemoteConfirmOpen(false);
                                                    setRemoteMasterPw('');
                                                }}
                                            />
                                            <div className="sheet sheetRemoteConfirm">
                                                <div className="sheetHdr">
                                                    <div className="sheetTitle">{tr('changeRepoTitle')}</div>
                                                    <button
                                                        className="xbtn"
                                                        type="button"
                                                        disabled={remoteBusy}
                                                        onClick={() => {
                                                            setRemoteConfirmOpen(false);
                                                            setRemoteMasterPw('');
                                                        }}
                                                        aria-label={tr('close')}
                                                    >
                                                        ×
                                                    </button>
                                                </div>
                                                <p className="accountRemoteConfirmTxt">{tr('changeRepoBody')}</p>
                                                <p className="accountRemoteConfirmTarget">
                                                    <span className="accountRemoteCurrentLbl">{tr('newTarget')}</span>{' '}
                                                    {remoteRepoURL.trim()}
                                                    {remoteBranch.trim() ? ` @${remoteBranch.trim()}` : ''}
                                                </p>
                                                <PasswordInput
                                                    inputClassName="inp"
                                                    wrapClassName="accountPatField"
                                                    placeholder={tr('masterPassword')}
                                                    value={remoteMasterPw}
                                                    onChange={(e) => setRemoteMasterPw(e.target.value)}
                                                    autoComplete="off"
                                                    disabled={remoteBusy}
                                                    locale={appLocale}
                                                    onKeyDown={(e) => {
                                                        if (e.key === 'Enter' && remoteMasterPw.trim() && !remoteBusy) {
                                                            e.preventDefault();
                                                            void doSaveRemote(remoteMasterPw);
                                                        }
                                                    }}
                                                />
                                                {accountErr && remoteConfirmOpen && (
                                                    <div className="err sheetBanner" role="alert">
                                                        {accountErr}
                                                    </div>
                                                )}
                                                <div className="accountRemoteConfirmActions">
                                                    <button
                                                        className="btn primary"
                                                        type="button"
                                                        disabled={remoteBusy || !remoteMasterPw.trim()}
                                                        onClick={() => void doSaveRemote(remoteMasterPw)}
                                                    >
                                                        {remoteBusy ? tr('copying') : tr('confirm')}
                                                    </button>
                                                    <button
                                                        className="btn"
                                                        type="button"
                                                        disabled={remoteBusy}
                                                        onClick={() => {
                                                            setRemoteConfirmOpen(false);
                                                            setRemoteMasterPw('');
                                                            setAccountErr('');
                                                        }}
                                                    >
                                                        {tr('cancel')}
                                                    </button>
                                                </div>
                                            </div>
                                        </div>
                                    )}

                                    <div className="row settings accountExportBlock">
                                        <label className="lbl inline">{tr('export')}</label>
                                        <div className="authImportList">
                                            <div className="authImportRow">
                                                <button
                                                    className="btn authImportBtn"
                                                    type="button"
                                                    onClick={doExportLocalVault}
                                                >
                                                    {tr('exportVault')}
                                                </button>
                                                <span className="authImportHint">{tr('exportVaultHint')}</span>
                                            </div>
                                            {!session?.localOnly && (
                                                <div className="authImportRow">
                                                    <button
                                                        className="btn authImportBtn"
                                                        type="button"
                                                        onClick={doExportFile}
                                                    >
                                                        {tr('exportProfile')}
                                                    </button>
                                                    <span className="authImportHint">{tr('exportProfileHint')}</span>
                                                </div>
                                            )}
                                        </div>
                                    </div>

                                    <div className="row settings accountSheetFooter">
                                        <button
                                            className="btn"
                                            type="button"
                                            onClick={() => {
                                                setAccountErr('');
                                                setAccountOpen(false);
                                            }}
                                        >
                                            {tr('back')}
                                        </button>
                                        <button
                                            className="btn danger"
                                            type="button"
                                            title={tr('deleteProfileTitle')}
                                            onClick={() => void doDeleteProfile()}
                                        >
                                            {tr('deleteProfileLocal')}
                                        </button>
                                    </div>
                                </div>
                            </div>
                        )}

                        <div className="vaultBoard">
                            <div className="vaultBoardMain" onClick={clearFolderSelectionIfBlank}>
                            <div className="folders">
                                {folderCreateOpen ? (
                                    <div className="fRowCreate">
                                        <input
                                            ref={newFolderInputRef}
                                            className="einp folderCreateInp"
                                            value={newFolderName}
                                            placeholder={tr('folderNamePlaceholder')}
                                            autoComplete="off"
                                            onChange={(e) => setNewFolderName(e.target.value)}
                                            onKeyDown={(e) => {
                                                if (e.key === 'Enter') {
                                                    e.preventDefault();
                                                    void doAddFolder();
                                                } else if (e.key === 'Escape') {
                                                    e.preventDefault();
                                                    cancelFolderCreate();
                                                }
                                            }}
                                        />
                                        <button
                                            className="iconBtn ok"
                                            type="button"
                                            title={tr('create')}
                                            disabled={!newFolderName.trim()}
                                            onClick={() => void doAddFolder()}
                                        >
                                            <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true">
                                                <path
                                                    d="M9 16.2 4.8 12 3.4 13.4 9 19 21 7 19.6 5.6z"
                                                    fill="currentColor"
                                                />
                                            </svg>
                                        </button>
                                        <button
                                            className="iconBtn"
                                            type="button"
                                            title={tr('cancel')}
                                            onClick={cancelFolderCreate}
                                        >
                                            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" aria-hidden="true">
                                                <path
                                                    d="M8 8l8 8M16 8l-8 8"
                                                    stroke="currentColor"
                                                    strokeWidth="2"
                                                    strokeLinecap="round"
                                                />
                                            </svg>
                                        </button>
                                    </div>
                                ) : (
                                    <button className="fbtn add" type="button" onClick={openFolderCreate}>
                                        {tr('newFolder')}
                                    </button>
                                )}
                                <div
                                    className="fList"
                                    onClick={(e) => {
                                        if (e.target === e.currentTarget) clearFolderSelection();
                                    }}
                                >
                                    {vault.folders?.map((f, idx) =>
                                        folderRenameId === f.id ? (
                                            <div key={f.id} className="fRowCreate" data-idx={idx}>
                                                <input
                                                    ref={folderRenameInputRef}
                                                    className="einp folderCreateInp"
                                                    value={folderRenameName}
                                                    placeholder={tr('folderNamePlaceholder')}
                                                    autoComplete="off"
                                                    onChange={(e) => setFolderRenameName(e.target.value)}
                                                    onKeyDown={(e) => {
                                                        if (e.key === 'Enter') {
                                                            e.preventDefault();
                                                            void doRenameFolder();
                                                        } else if (e.key === 'Escape') {
                                                            e.preventDefault();
                                                            cancelFolderRename();
                                                        }
                                                    }}
                                                />
                                                <button
                                                    className="iconBtn ok"
                                                    type="button"
                                                    title={tr('save')}
                                                    disabled={!folderRenameName.trim()}
                                                    onClick={() => void doRenameFolder()}
                                                >
                                                    <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true">
                                                        <path
                                                            d="M9 16.2 4.8 12 3.4 13.4 9 19 21 7 19.6 5.6z"
                                                            fill="currentColor"
                                                        />
                                                    </svg>
                                                </button>
                                                <button
                                                    className="iconBtn"
                                                    type="button"
                                                    title={tr('cancel')}
                                                    onClick={cancelFolderRename}
                                                >
                                                    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" aria-hidden="true">
                                                        <path
                                                            d="M8 8l8 8M16 8l-8 8"
                                                            stroke="currentColor"
                                                            strokeWidth="2"
                                                            strokeLinecap="round"
                                                        />
                                                    </svg>
                                                </button>
                                            </div>
                                        ) : (
                                        <div
                                            key={f.id}
                                            className="fRow"
                                            data-idx={idx}
                                            onContextMenu={(e) => {
                                                e.preventDefault();
                                                openFolderMenuAt(f.id, e.clientX, e.clientY);
                                            }}
                                        >
                                            <button
                                                type="button"
                                                className={f.id === folderId ? 'fbtn active' : 'fbtn'}
                                                title={f.name}
                                                onClick={() => {
                                                    closeFolderMenu();
                                                    if (f.id === folderId) {
                                                        clearFolderSelection();
                                                        return;
                                                    }
                                                    setFolderId(f.id);
                                                    setNewDescTitle('');
                                                    setNewDescKey('');
                                                    setNewDescPassword('');
                                                    setNewDescVal('');
                                                    setDescModalMode(null);
                                                }}
                                            >
                                                <span className="fbtnLabel">{f.name}</span>
                                            </button>
                                            <button
                                                type="button"
                                                className="fbtnMore"
                                                title={tr('folderActions')}
                                                aria-label={tr('folderActions')}
                                                aria-haspopup="menu"
                                                aria-expanded={folderMenu?.folderId === f.id}
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    const r = e.currentTarget.getBoundingClientRect();
                                                    openFolderMenuAt(f.id, r.left, r.bottom + 4);
                                                }}
                                            >
                                                ⋯
                                            </button>
                                        </div>
                                        )
                                    )}
                                </div>
                            </div>

                            <div
                                className="descPane"
                                onClick={(e) => {
                                    if (e.target === e.currentTarget) clearFolderSelection();
                                }}
                            >
                                {!folderId ? (
                                    <p className="descPaneHint">{tr('selectFolderHint')}</p>
                                ) : (
                                    <>
                                        <button className="fbtn add descNewOpenBtn" type="button" onClick={openDescNewModal}>
                                            {tr('newRecord')}
                                        </button>
                                        <div
                                            className="dList"
                                            onClick={(e) => {
                                                if (e.target === e.currentTarget) clearFolderSelection();
                                            }}
                                        >
                                            {descriptionsSorted.map((d) => (
                                                <div key={d.id} className="dRow">
                                                    <button
                                                        type="button"
                                                        className={
                                                            d.id === selectedDescId && descModalMode === 'edit'
                                                                ? 'dbtn active'
                                                                : 'dbtn'
                                                        }
                                                        title={d.title}
                                                        onClick={() => openDescEditModal(d)}
                                                    >
                                                        <span className="dbtnLabel">{d.title}</span>
                                                    </button>
                                                    <button
                                                        className="iconBtn danger dtrash"
                                                        type="button"
                                                        title={tr('deleteRecord')}
                                                        onClick={() => void doDeleteDescription(d.id)}
                                                    >
                                                        <svg viewBox="0 0 24 24" width="18" height="18" fill="none" aria-hidden="true">
                                                            <path
                                                                d="M8 8l8 8M16 8l-8 8"
                                                                stroke="currentColor"
                                                                strokeWidth="2"
                                                                strokeLinecap="round"
                                                            />
                                                        </svg>
                                                    </button>
                                                </div>
                                            ))}
                                        </div>
                                    </>
                                )}
                            </div>
                            </div>

                        {folderId && descModalMode && (
                            <div className="modal descVaultModal" role="dialog" aria-modal="true">
                                <div className="backdrop" onClick={() => tryCloseDescModal()} />
                                <div className="sheet sheetDesc">
                                    <div className="sheetHdr">
                                        <div className="sheetTitle">
                                            {descModalMode === 'new' ? tr('newRecordTitle') : ''}
                                        </div>
                                        <button
                                            className="xbtn"
                                            type="button"
                                            onClick={() => tryCloseDescModal()}
                                            aria-label={tr('close')}
                                        >
                                            ×
                                        </button>
                                    </div>

                                    {descModalMode === 'new' && (
                                        <div className="descModalNew">
                                            <input
                                                className="einp descTitleInp"
                                                placeholder={tr('fieldTitle')}
                                                value={newDescTitle}
                                                onChange={(e) => setNewDescTitle(e.target.value)}
                                            />
                                            <input
                                                className="einp key"
                                                placeholder={tr('fieldLogin')}
                                                value={newDescKey}
                                                onChange={(e) => setNewDescKey(e.target.value)}
                                            />
                                            <input
                                                className="einp key"
                                                placeholder={tr('fieldPassword')}
                                                value={newDescPassword}
                                                onChange={(e) => setNewDescPassword(e.target.value)}
                                                autoComplete="off"
                                            />
                                            <div className="descNotesBlock">
                                                <label className="descNotesLbl">
                                                    {tr('fieldNotes')}
                                                    <span className="fieldHint">{tr('notesExpandHint')}</span>
                                                </label>
                                                <div className="valWrap valWrapVal">
                                                    <textarea
                                                        className="einp val einpValTex"
                                                        data-expanded={expandedValIds.has(NEW_DESC_VAL_EXPAND_ID) ? '1' : '0'}
                                                        placeholder={tr('fieldNotes')}
                                                        rows={expandedValIds.has(NEW_DESC_VAL_EXPAND_ID) ? 12 : 1}
                                                        value={newDescVal}
                                                        onChange={(e) => setNewDescVal(e.target.value)}
                                                        onDoubleClick={() => toggleExpandValue(NEW_DESC_VAL_EXPAND_ID)}
                                                        title={tr('dblClickExpandTitle')}
                                                    />
                                                </div>
                                            </div>
                                            <div className="descModalFooter">
                                                <button
                                                    className="btn primary"
                                                    type="button"
                                                    onClick={() => void doAddDescription()}
                                                    disabled={!newDescTitle.trim()}
                                                >
                                                    {tr('save')}
                                                </button>
                                                <button className="btn" type="button" onClick={() => tryCloseDescModal()}>
                                                    {tr('cancel')}
                                                </button>
                                            </div>
                                        </div>
                                    )}

                                    {descModalMode === 'edit' && selectedDesc && (
                                        <>
                                            <DescDetailPanel
                                                desc={selectedDesc}
                                                rootClassName="descDetail--inModal"
                                                locale={appLocale}
                                                titleDraft={titleDraft}
                                                keyDraft={keyDraft}
                                                passwordDraft={passwordDraft}
                                                valueDraft={valueDraft}
                                                expandedValIds={expandedValIds}
                                                onTitleChange={(v) =>
                                                    setTitleDraft((p) => ({ ...p, [selectedDesc.id]: v }))
                                                }
                                                onKeyChange={(v) =>
                                                    setKeyDraft((p) => ({ ...p, [selectedDesc.id]: v }))
                                                }
                                                onPasswordChange={(v) =>
                                                    setPasswordDraft((p) => ({ ...p, [selectedDesc.id]: v }))
                                                }
                                                onValueChange={(v) =>
                                                    setValueDraft((p) => ({ ...p, [selectedDesc.id]: v }))
                                                }
                                                onCopy={(text) => void copyEntryValue(text)}
                                                onToggleExpandValue={() => toggleExpandValue(selectedDesc.id)}
                                                onApplyTitle={() => void applyTitleDraft(selectedDesc.id)}
                                                onApplyKey={() => void applyKeyDraft(selectedDesc.id)}
                                                onApplyPassword={() => void applyPasswordDraft(selectedDesc.id)}
                                                onApplyValue={() => void applyValueDraft(selectedDesc.id)}
                                                onDiscardTitle={() => discardTitleDraft(selectedDesc.id)}
                                                onDiscardKey={() => discardKeyDraft(selectedDesc.id)}
                                                onDiscardPassword={() => discardPasswordDraft(selectedDesc.id)}
                                                onDiscardValue={() => discardValueDraft(selectedDesc.id)}
                                            />
                                            <div className="descModalFooter">
                                                <button className="btn" type="button" onClick={() => tryCloseDescModal()}>
                                                    {tr('close')}
                                                </button>
                                            </div>
                                        </>
                                    )}
                                </div>
                            </div>
                        )}

                        {folderMenu && (
                            <>
                                <div className="folderMenuBackdrop" aria-hidden="true" onClick={closeFolderMenu} />
                                <div
                                    className="folderMenu"
                                    role="menu"
                                    style={{ left: folderMenu.x, top: folderMenu.y }}
                                >
                                    <button
                                        type="button"
                                        className="folderMenuItem"
                                        role="menuitem"
                                        onClick={() => {
                                            const id = folderMenu.folderId;
                                            closeFolderMenu();
                                            startFolderRename(id);
                                        }}
                                    >
                                        {tr('rename')}
                                    </button>
                                    <button
                                        type="button"
                                        className="folderMenuItem folderMenuItemDanger"
                                        role="menuitem"
                                        onClick={() => {
                                            const id = folderMenu.folderId;
                                            closeFolderMenu();
                                            void doDeleteFolder(id);
                                        }}
                                    >
                                        {tr('delete')}
                                    </button>
                                </div>
                            </>
                        )}
                        </div>
                    </div>
                </main>
            )}
        </div>
    );
}
