'use client';

import { CircleCheck } from 'lucide-react';
import Button from './Button';
import { FONT, INK, type TipsTheme } from './tipsData';

export default function TipSubmitted({
  theme,
  name,
  onGoHome,
  onShareAnother,
}: {
  theme: TipsTheme;
  name?: string | null;
  onGoHome: () => void;
  onShareAnother: () => void;
}) {
  const greeting = name ? `Thank you, ${name}` : 'Thank you';

  return (
    <div
      style={{
        flex: 1,
        display: 'flex',
        flexDirection: 'column',
        minHeight: 0,
      }}
    >
      {/* Centered confirmation */}
      <div
        style={{
          flex: 1,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '24px 24px 0',
          gap: 24,
        }}
      >
        <CircleCheck
          size={88}
          color={theme.successColor}
          strokeWidth={2}
          aria-hidden
        />
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            gap: 12,
          }}
        >
          <h1
            style={{
              margin: 0,
              fontFamily: FONT,
              fontWeight: 700,
              fontSize: 26,
              lineHeight: '34px',
              letterSpacing: '-0.3px',
              color: INK,
              textAlign: 'center',
            }}
          >
            {greeting}
            <br />
            your tip is being reviewed.
          </h1>
          <p
            style={{
              margin: 0,
              maxWidth: 320,
              fontFamily: FONT,
              fontWeight: 400,
              fontSize: 15,
              lineHeight: '22px',
              color: 'var(--color-grey-600)',
              textAlign: 'center',
            }}
          >
            Your tip is being reviewed. We&apos;ll notify you once it&apos;s
            live, usually within 24 hours.
          </p>
        </div>
      </div>

      {/* Pinned footer */}
      <div
        style={{
          flexShrink: 0,
          borderTop: '1px solid var(--color-grey-100)',
          padding: '16px 16px calc(20px + var(--rn-sab, env(safe-area-inset-bottom)))',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          gap: 12,
          backgroundColor: 'var(--color-grey-0)',
        }}
      >
        <Button label="Go to Home" variant="dark" size="xl" fullWidth onClick={onGoHome} />
        <Button label="Share another tip" variant="noFill" size="lg" onClick={onShareAnother} />
      </div>
    </div>
  );
}
