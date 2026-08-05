import type { ReactNode } from 'react';
import { BrowserOpenURL } from '../wailsjs/runtime/runtime';
import { t, type MsgKey } from './i18n/messages';
import type { AppLocale } from './locale';

const GITHUB_TOKENS_URL = 'https://github.com/settings/tokens?type=beta';
const GIT_VERSION_CMD = 'git --version';

type CreateProfileHelpProps = {
    open: boolean;
    onClose: () => void;
    locale: AppLocale;
};

/** Вставляет `marker` в виде React-узла внутрь отформатированной строки (для code/ссылок в переводах). */
function withMarkers(text: string, markers: Record<string, ReactNode>): ReactNode {
    const keys = Object.keys(markers).filter((k) => k.length > 0);
    let bestKey = '';
    let bestIdx = -1;
    for (const k of keys) {
        const idx = text.indexOf(k);
        if (idx !== -1 && (bestIdx === -1 || idx < bestIdx)) {
            bestIdx = idx;
            bestKey = k;
        }
    }
    if (bestIdx === -1) return text;
    const before = text.slice(0, bestIdx);
    const after = text.slice(bestIdx + bestKey.length);
    const rest = { ...markers };
    delete rest[bestKey];
    return (
        <>
            {before}
            {markers[bestKey]}
            {withMarkers(after, rest)}
        </>
    );
}

export function CreateProfileHelp({ open, onClose, locale }: CreateProfileHelpProps) {
    if (!open) return null;

    const tr = (key: MsgKey, vars?: Record<string, string | number>) => t(locale, key, vars);
    const exampleRepoUrl =
        locale === 'ru' ? 'https://github.com/ВАШ_ЛОГИН/ИМЯ_РЕПО.git' : 'https://github.com/YOUR_USERNAME/REPO_NAME.git';

    return (
        <div className="modal createHelpModal" role="dialog" aria-modal="true" aria-labelledby="createHelpTitle">
            <div className="backdrop" onClick={onClose} />
            <div className="sheet sheetCreateHelp">
                <div className="sheetHdr">
                    <div className="sheetTitle" id="createHelpTitle">
                        {tr('helpTitle')}
                    </div>
                    <button className="xbtn" type="button" onClick={onClose} aria-label={tr('close')}>
                        ×
                    </button>
                </div>

                <div className="createHelpBody">
                    <p>{tr('helpIntro')}</p>
                    <ol className="createHelpList">
                        <li>{tr('helpOpt1')}</li>
                        <li>{tr('helpOpt2')}</li>
                    </ol>
                    <p>{tr('helpEachHasProsCons')}</p>

                    <section className="createHelpSection">
                        <h3 className="createHelpH">{tr('helpLocalTitle')}</h3>
                        <p>{tr('helpLocalBody')}</p>
                        <h4 className="createHelpSubH">{tr('helpLocalProsTitle')}</h4>
                        <ul className="createHelpList">
                            <li>{tr('helpLocalPro1')}</li>
                            <li>{tr('helpLocalPro2')}</li>
                        </ul>
                        <h4 className="createHelpSubH">{tr('helpConsTitle')}</h4>
                        <ul className="createHelpList">
                            <li>{tr('helpLocalCon1')}</li>
                            <li>{tr('helpLocalCon2')}</li>
                        </ul>
                    </section>

                    <section className="createHelpSection">
                        <h3 className="createHelpH">{tr('helpGitTitle')}</h3>
                        <p>{tr('helpGitBody')}</p>
                        <h4 className="createHelpSubH">{tr('helpProsTitle')}</h4>
                        <ul className="createHelpList">
                            <li>{tr('helpGitPro1')}</li>
                            <li>{tr('helpGitPro2')}</li>
                        </ul>
                        <h4 className="createHelpSubH">{tr('helpConsTitle')}</h4>
                        <ul className="createHelpList">
                            <li>{tr('helpGitCon1')}</li>
                            <li>{tr('helpGitCon2')}</li>
                        </ul>
                    </section>

                    <section className="createHelpSection">
                        <h3 className="createHelpH">{tr('helpGitSetupTitle')}</h3>

                        <div className="createHelpStep">
                            <span className="createHelpStepNum" aria-hidden="true">
                                1
                            </span>
                            <div className="createHelpStepBody">
                                <p>
                                    {tr('helpStep1Title')}
                                    <br />
                                    {withMarkers(tr('helpStep1Body', { cmd: GIT_VERSION_CMD }), {
                                        [GIT_VERSION_CMD]: (
                                            <code className="createHelpCode">{GIT_VERSION_CMD}</code>
                                        ),
                                    })}
                                </p>
                            </div>
                        </div>

                        <div className="createHelpStep">
                            <span className="createHelpStepNum" aria-hidden="true">
                                2
                            </span>
                            <div className="createHelpStepBody">
                                <p>{tr('helpStep2Title')}</p>
                                <ul className="createHelpList">
                                    <li>
                                        {withMarkers(tr('helpStep2Li1', { main: 'main', master: 'master' }), {
                                            main: <code className="createHelpCode">main</code>,
                                            master: <code className="createHelpCode">master</code>,
                                        })}
                                    </li>
                                    <li>{tr('helpStep2Li2')}</li>
                                    <li>
                                        {withMarkers(tr('helpStep2Li3', { url: exampleRepoUrl }), {
                                            [exampleRepoUrl]: (
                                                <code className="createHelpCode">{exampleRepoUrl}</code>
                                            ),
                                        })}
                                    </li>
                                </ul>
                            </div>
                        </div>

                        <div className="createHelpStep">
                            <span className="createHelpStepNum" aria-hidden="true">
                                3
                            </span>
                            <div className="createHelpStepBody">
                                <p>{tr('helpStep3Title')}</p>
                                <ul className="createHelpList">
                                    <li>{tr('helpStep3Li1')}</li>
                                    <li>
                                        {tr('helpStep3Li2')}
                                        <br />
                                        {withMarkers(tr('helpStep3DirectLink', { url: GITHUB_TOKENS_URL }), {
                                            [GITHUB_TOKENS_URL]: (
                                                <button
                                                    type="button"
                                                    className="createHelpLink"
                                                    onClick={() => BrowserOpenURL(GITHUB_TOKENS_URL)}
                                                >
                                                    {GITHUB_TOKENS_URL}
                                                </button>
                                            ),
                                        })}
                                    </li>
                                    <li>{tr('helpStep3Li3')}</li>
                                    <li>
                                        {tr('helpStep3Fill')}
                                        <ul className="createHelpList createHelpListNested">
                                            <li>{tr('helpStep3Name')}</li>
                                            <li>{tr('helpStep3Exp')}</li>
                                            <li>{tr('helpStep3Owner')}</li>
                                            <li>{tr('helpStep3Access')}</li>
                                        </ul>
                                    </li>
                                    <li>
                                        {tr('helpStep3Perms')}
                                        <ul className="createHelpList createHelpListNested">
                                            <li>{tr('helpStep3Contents')}</li>
                                            <li>{tr('helpStep3Meta')}</li>
                                        </ul>
                                        {tr('helpStep3OtherPerms')}
                                    </li>
                                    <li>{tr('helpStep3Generate')}</li>
                                    <li>{tr('helpStep3Paste')}</li>
                                </ul>
                            </div>
                        </div>
                    </section>
                </div>
            </div>
        </div>
    );
}
