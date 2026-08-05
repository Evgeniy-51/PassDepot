import type { vaultcore } from '../wailsjs/go/models';
import { t, type MsgKey } from './i18n/messages';
import type { AppLocale } from './locale';

export type DescDetailPanelProps = {
    desc: vaultcore.Description;
    /** Доп. классы на корень (например descDetail--inModal) */
    rootClassName?: string;
    locale: AppLocale;
    titleDraft: Record<string, string>;
    keyDraft: Record<string, string>;
    passwordDraft: Record<string, string>;
    valueDraft: Record<string, string>;
    expandedValIds: Set<string>;
    onTitleChange: (value: string) => void;
    onKeyChange: (value: string) => void;
    onPasswordChange: (value: string) => void;
    onValueChange: (value: string) => void;
    onCopy: (text: string) => void;
    onToggleExpandValue: () => void;
    onApplyTitle: () => void;
    onApplyKey: () => void;
    onApplyPassword: () => void;
    onApplyValue: () => void;
    onDiscardTitle: () => void;
    onDiscardKey: () => void;
    onDiscardPassword: () => void;
    onDiscardValue: () => void;
};

function CopyIcon() {
    return (
        <svg viewBox="0 0 24 24" width="16" height="16">
            <path
                d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"
                fill="currentColor"
            />
        </svg>
    );
}

function OkIcon() {
    return (
        <svg viewBox="0 0 24 24" width="18" height="18">
            <path d="M9 16.2 4.8 12 3.4 13.4 9 19 21 7 19.6 5.6z" fill="currentColor" />
        </svg>
    );
}

function CancelIcon() {
    return (
        <svg viewBox="0 0 24 24" width="18" height="18" fill="none" aria-hidden="true">
            <path
                d="M8 8l8 8M16 8l-8 8"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
            />
        </svg>
    );
}

export function DescDetailPanel({
    desc,
    rootClassName,
    locale,
    titleDraft,
    keyDraft,
    passwordDraft,
    valueDraft,
    expandedValIds,
    onTitleChange,
    onKeyChange,
    onPasswordChange,
    onValueChange,
    onCopy,
    onToggleExpandValue,
    onApplyTitle,
    onApplyKey,
    onApplyPassword,
    onApplyValue,
    onDiscardTitle,
    onDiscardKey,
    onDiscardPassword,
    onDiscardValue,
}: DescDetailPanelProps) {
    const tr = (key: MsgKey) => t(locale, key);
    const sid = desc.id;
    const titleShown = titleDraft[sid] ?? desc.title;
    const keyShown = keyDraft[sid] ?? desc.key;
    const passwordShown = passwordDraft[sid] ?? desc.password ?? '';
    const valShown = valueDraft[sid] ?? desc.value;
    const isTitleDraft = titleShown !== desc.title;
    const isKeyDraft = keyShown !== desc.key;
    const isPasswordDraft = passwordShown !== (desc.password ?? '');
    const isValDraft = valShown !== desc.value;

    return (
        <div className={'descDetail' + (rootClassName ? ` ${rootClassName}` : '')}>
            <div className={'descDetailField' + (isTitleDraft ? ' descDetailFieldDraft' : '')}>
                <div className="valWrap valWrapKey">
                    <input
                        className="einp"
                        value={titleShown}
                        aria-label={tr('fieldTitle')}
                        placeholder={tr('fieldTitle')}
                        onChange={(ev) => onTitleChange(ev.target.value)}
                    />
                    <button className="iconBtn mini" type="button" title={tr('copy')} onClick={() => onCopy(titleShown)}>
                        <CopyIcon />
                    </button>
                </div>
                <div className="descFieldActions">
                    {isTitleDraft && (
                        <>
                            <button className="iconBtn ok" type="button" title={tr('writeToVault')} onClick={() => void onApplyTitle()}>
                                <OkIcon />
                            </button>
                            <button className="iconBtn" type="button" title={tr('cancelVerb')} onClick={onDiscardTitle}>
                                <CancelIcon />
                            </button>
                        </>
                    )}
                </div>
            </div>
            <div className={'descDetailField' + (isKeyDraft ? ' descDetailFieldDraft' : '')}>
                <div className="valWrap valWrapKey">
                    <input
                        className="einp key"
                        value={keyShown}
                        aria-label={tr('fieldLogin')}
                        placeholder={tr('fieldLogin')}
                        onChange={(ev) => onKeyChange(ev.target.value)}
                    />
                    <button className="iconBtn mini" type="button" title={tr('copy')} onClick={() => onCopy(keyShown)}>
                        <CopyIcon />
                    </button>
                </div>
                <div className="descFieldActions">
                    {isKeyDraft && (
                        <>
                            <button className="iconBtn ok" type="button" title={tr('writeToVault')} onClick={() => void onApplyKey()}>
                                <OkIcon />
                            </button>
                            <button className="iconBtn" type="button" title={tr('cancelVerb')} onClick={onDiscardKey}>
                                <CancelIcon />
                            </button>
                        </>
                    )}
                </div>
            </div>
            <div className={'descDetailField' + (isPasswordDraft ? ' descDetailFieldDraft' : '')}>
                <div className="valWrap valWrapKey">
                    <input
                        className="einp key"
                        value={passwordShown}
                        aria-label={tr('fieldPassword')}
                        placeholder={tr('fieldPassword')}
                        autoComplete="off"
                        onChange={(ev) => onPasswordChange(ev.target.value)}
                    />
                    <button className="iconBtn mini" type="button" title={tr('copy')} onClick={() => onCopy(passwordShown)}>
                        <CopyIcon />
                    </button>
                </div>
                <div className="descFieldActions">
                    {isPasswordDraft && (
                        <>
                            <button
                                className="iconBtn ok"
                                type="button"
                                title={tr('writeToVault')}
                                onClick={() => void onApplyPassword()}
                            >
                                <OkIcon />
                            </button>
                            <button className="iconBtn" type="button" title={tr('cancelVerb')} onClick={onDiscardPassword}>
                                <CancelIcon />
                            </button>
                        </>
                    )}
                </div>
            </div>
            <div className={'descDetailField descDetailFieldVal' + (isValDraft ? ' descDetailFieldDraft' : '')}>
                <div className="descNotesLbl">
                    {tr('fieldNotes')}
                    <span className="fieldHint">{tr('notesExpandHint')}</span>
                </div>
                <div className="valWrap valWrapVal">
                    <textarea
                        className="einp val einpValTex"
                        data-expanded={expandedValIds.has(sid) ? '1' : '0'}
                        rows={expandedValIds.has(sid) ? 12 : 1}
                        value={valShown}
                        aria-label={tr('fieldNotes')}
                        placeholder={tr('fieldNotes')}
                        onChange={(ev) => onValueChange(ev.target.value)}
                        onDoubleClick={onToggleExpandValue}
                        title={tr('dblClickExpandTitle')}
                    />
                    <button className="iconBtn mini" type="button" title={tr('copy')} onClick={() => onCopy(valShown)}>
                        <CopyIcon />
                    </button>
                </div>
                <div className="descFieldActions">
                    {isValDraft && (
                        <>
                            <button className="iconBtn ok" type="button" title={tr('writeToVault')} onClick={() => void onApplyValue()}>
                                <OkIcon />
                            </button>
                            <button className="iconBtn" type="button" title={tr('cancelVerb')} onClick={onDiscardValue}>
                                <CancelIcon />
                            </button>
                        </>
                    )}
                </div>
            </div>
        </div>
    );
}
