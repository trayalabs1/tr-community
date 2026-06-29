'use client';

import {
  CARD_BORDER,
  FONT,
  GREY,
  INK,
  TOPICS,
  type TipTopic,
  type TipsTheme,
} from './tipsData';

export default function TopicSelection({
  theme,
  selectedId,
  onSelect,
}: {
  theme: TipsTheme;
  selectedId: string | null;
  onSelect: (id: string) => void;
}) {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'flex-start',
        gap: 16,
        padding: '0 16px 20px',
      }}
    >
      <h1
        style={{
          margin: 0,
          alignSelf: 'stretch',
          fontFamily: FONT,
          fontWeight: 600,
          fontSize: 26,
          lineHeight: '30px',
          letterSpacing: '-0.26px',
          color: INK,
        }}
      >
        Your story could help someone in their first month
      </h1>

      <p
        style={{
          margin: 0,
          alignSelf: 'stretch',
          fontFamily: FONT,
          fontWeight: 500,
          fontSize: 16,
          lineHeight: '20px',
          color: GREY[650],
        }}
      >
        Select topic
      </p>

      {/* Figma "Frame 2147231584": row · wrap · flex-start · gap 16. */}
      <div
        style={{
          display: 'flex',
          flexDirection: 'row',
          flexWrap: 'wrap',
          alignItems: 'flex-start',
          alignContent: 'flex-start',
          gap: 16,
          alignSelf: 'stretch',
        }}
      >
        {TOPICS.map((topic) => (
          <TopicCard
            key={topic.id}
            topic={topic}
            theme={theme}
            selected={selectedId === topic.id}
            onClick={() => onSelect(topic.id)}
          />
        ))}
      </div>
    </div>
  );
}

function TopicCard({
  topic,
  theme,
  selected,
  onClick,
}: {
  topic: TipTopic;
  theme: TipsTheme;
  selected: boolean;
  onClick: () => void;
}) {
  const { Icon } = topic;
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={selected}
      style={{
        // Two per row: fill the available width but keep the ~156px design ratio.
        position: 'relative',
        flex: '1 1 140px',
        minWidth: 140,
        minHeight: 120,
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'flex-end',
        alignItems: 'flex-start',
        boxSizing: 'border-box',
        padding: 12,
        borderRadius: 12,
        cursor: 'pointer',
        textAlign: 'left',
        backgroundColor: selected ? theme.selectedFill : GREY[0],
        border: `1px solid ${selected ? theme.selectedBorder : CARD_BORDER}`,
        transition: 'background-color 150ms ease, border-color 150ms ease',
        fontFamily: FONT,
      }}
    >
      {/* Radio pinned top-right so the title can use the full card width. */}
      <span style={{ position: 'absolute', top: 12, right: 12 }}>
        <Radio selected={selected} color={theme.radio} />
      </span>

      <span
        style={{
          width: '100%',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'flex-start',
          gap: 12,
          paddingTop: 20,
        }}
      >
        <Icon size={24} color="#404040" strokeWidth={1.5} />
        <span
          style={{
            fontWeight: 600,
            fontSize: 16,
            lineHeight: '20px',
            color: '#404040',
          }}
        >
          {topic.title}
        </span>
      </span>
    </button>
  );
}

function Radio({ selected, color }: { selected: boolean; color: string }) {
  return (
    <span
      aria-hidden
      style={{
        flexShrink: 0,
        width: 24,
        height: 24,
        borderRadius: '50%',
        boxSizing: 'border-box',
        border: `2px solid ${selected ? color : GREY[200]}`,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        backgroundColor: 'transparent',
      }}
    >
      {selected && (
        <span
          style={{
            width: 12,
            height: 12,
            borderRadius: '50%',
            backgroundColor: color,
          }}
        />
      )}
    </span>
  );
}
