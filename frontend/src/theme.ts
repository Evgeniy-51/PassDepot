import { WindowSetDarkTheme, WindowSetLightTheme } from '../wailsjs/runtime/runtime';

export type AppTheme = 'blue' | 'coral' | 'light';

export const THEME_ORDER: AppTheme[] = ['blue', 'coral', 'light'];

export const THEME_LABEL: Record<AppTheme, string> = {
    blue: 'Blue',
    coral: 'Coral',
    light: 'Light',
};

const STORAGE_KEY = 'passdepot.theme';

export function isAppTheme(v: unknown): v is AppTheme {
    return v === 'blue' || v === 'coral' || v === 'light';
}

export function readStoredTheme(): AppTheme {
    try {
        const raw = localStorage.getItem(STORAGE_KEY);
        if (raw === 'red') return 'coral'; // legacy id
        if (isAppTheme(raw)) return raw;
    } catch {
        /* ignore */
    }
    return 'blue';
}

function syncWindowChrome(theme: AppTheme) {
    try {
        if (theme === 'light') WindowSetLightTheme();
        else WindowSetDarkTheme();
    } catch {
        /* runtime may be unavailable outside Wails */
    }
}

/** Apply theme to <html> and persist. */
export function applyTheme(theme: AppTheme) {
    document.documentElement.dataset.theme = theme;
    try {
        localStorage.setItem(STORAGE_KEY, theme);
    } catch {
        /* ignore */
    }
    syncWindowChrome(theme);
}

export function applyStoredTheme() {
    applyTheme(readStoredTheme());
}

export function nextTheme(current: AppTheme): AppTheme {
    const i = THEME_ORDER.indexOf(current);
    return THEME_ORDER[(i + 1) % THEME_ORDER.length];
}

export function cycleTheme(): AppTheme {
    const next = nextTheme(readStoredTheme());
    applyTheme(next);
    return next;
}
