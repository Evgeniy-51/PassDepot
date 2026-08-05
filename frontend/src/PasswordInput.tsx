import { useState, type InputHTMLAttributes } from 'react';
import { t } from './i18n/messages';
import { readStoredLocale, type AppLocale } from './locale';

type PasswordInputProps = Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> & {
    inputClassName?: string;
    wrapClassName?: string;
    locale?: AppLocale;
};

function EyeIcon() {
    return (
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" aria-hidden="true">
            <path
                d="M2.25 12s3.75-6.75 9.75-6.75S21.75 12 21.75 12s-3.75 6.75-9.75 6.75S2.25 12 2.25 12Z"
                stroke="currentColor"
                strokeWidth="1.75"
                strokeLinejoin="round"
            />
            <circle cx="12" cy="12" r="2.75" stroke="currentColor" strokeWidth="1.75" />
        </svg>
    );
}

function EyeOffIcon() {
    return (
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" aria-hidden="true">
            <path
                d="M3 3l18 18M10.2 10.2A2.75 2.75 0 0 0 12 14.75c1.1 0 2.05-.65 2.48-1.58M6.53 6.74C4.62 8.04 3.1 9.94 2.25 12c0 0 3.75 6.75 9.75 6.75 1.55 0 2.98-.35 4.24-.96M15.9 9.9c.35.5.55 1.1.55 1.75 0 1.8-1.46 3.25-3.25 3.25-.65 0-1.25-.2-1.75-.55"
                stroke="currentColor"
                strokeWidth="1.75"
                strokeLinecap="round"
                strokeLinejoin="round"
            />
        </svg>
    );
}

export function PasswordInput({
    inputClassName = '',
    wrapClassName = '',
    disabled,
    locale = readStoredLocale(),
    ...props
}: PasswordInputProps) {
    const [visible, setVisible] = useState(false);

    return (
        <div className={`passWrap ${wrapClassName}`.trim()}>
            <input
                {...props}
                disabled={disabled}
                type={visible ? 'text' : 'password'}
                className={`passInp ${inputClassName}`.trim()}
            />
            <button
                type="button"
                className="passToggle"
                onClick={() => setVisible((v) => !v)}
                aria-label={visible ? t(locale, 'hidePassword') : t(locale, 'showPassword')}
                aria-pressed={visible}
                title={visible ? t(locale, 'hidePassword') : t(locale, 'showPassword')}
            >
                {visible ? <EyeOffIcon /> : <EyeIcon />}
            </button>
        </div>
    );
}
