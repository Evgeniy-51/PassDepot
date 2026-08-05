export type AppLocale = 'ru' | 'en';

export const LOCALE_ORDER: AppLocale[] = ['ru', 'en'];

export const LOCALE_LABEL: Record<AppLocale, string> = {
    ru: 'Русский',
    en: 'English',
};

const STORAGE_KEY = 'passdepot.locale';

export function isAppLocale(v: unknown): v is AppLocale {
    return v === 'ru' || v === 'en';
}

export function readStoredLocale(): AppLocale {
    try {
        const raw = localStorage.getItem(STORAGE_KEY);
        if (isAppLocale(raw)) return raw;
    } catch {
        /* ignore */
    }
    return 'ru';
}

/** Persist + data-locale / lang на <html>; язык UI через i18n + SetUILocale. */
export function applyLocale(locale: AppLocale) {
    document.documentElement.dataset.locale = locale;
    document.documentElement.lang = locale === 'ru' ? 'ru' : 'en';
    try {
        localStorage.setItem(STORAGE_KEY, locale);
    } catch {
        /* ignore */
    }
}

export function applyStoredLocale() {
    applyLocale(readStoredLocale());
}

export function nextLocale(current: AppLocale): AppLocale {
    const i = LOCALE_ORDER.indexOf(current);
    return LOCALE_ORDER[(i + 1) % LOCALE_ORDER.length];
}

export function cycleLocale(): AppLocale {
    const next = nextLocale(readStoredLocale());
    applyLocale(next);
    return next;
}
