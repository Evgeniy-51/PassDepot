type FlagProps = {
    className?: string;
    title?: string;
};

/** Пропорции 4:3; размер задаёт CSS (16×12 px). */
export function FlagRu({ className, title }: FlagProps) {
    return (
        <svg
            className={className}
            viewBox="0 0 12 9"
            width={16}
            height={12}
            aria-hidden={title ? undefined : true}
            role={title ? 'img' : undefined}
            focusable="false"
        >
            {title ? <title>{title}</title> : null}
            <rect width="12" height="9" rx="1" fill="#fff" />
            <rect y="0" width="12" height="3" fill="#ffffff" />
            <rect y="3" width="12" height="3" fill="#0039a6" />
            <rect y="6" width="12" height="3" fill="#d52b1e" />
            <rect width="12" height="9" rx="1" fill="none" stroke="rgba(0,0,0,0.18)" strokeWidth="0.5" />
        </svg>
    );
}

export function FlagGb({ className, title }: FlagProps) {
    return (
        <svg
            className={className}
            viewBox="0 0 12 9"
            width={16}
            height={12}
            aria-hidden={title ? undefined : true}
            role={title ? 'img' : undefined}
            focusable="false"
        >
            {title ? <title>{title}</title> : null}
            <rect width="12" height="9" rx="1" fill="#012169" />
            <path d="M0 0 L12 9 M12 0 L0 9" stroke="#fff" strokeWidth="1.8" />
            <path d="M0 0 L12 9" stroke="#c8102e" strokeWidth="1" />
            <path d="M12 0 L0 9" stroke="#c8102e" strokeWidth="1" />
            <path d="M6 0 V9 M0 4.5 H12" stroke="#fff" strokeWidth="3" />
            <path d="M6 0 V9 M0 4.5 H12" stroke="#c8102e" strokeWidth="1.6" />
            <rect width="12" height="9" rx="1" fill="none" stroke="rgba(0,0,0,0.18)" strokeWidth="0.5" />
        </svg>
    );
}
