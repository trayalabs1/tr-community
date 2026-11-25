'use client';

import { useRef, useState } from 'react';
import { Images, Send } from 'lucide-react';
import { FONT, GREY, INK, type TipTopic, type TipsTheme } from './tipsData';


export default function WriteTip({
  theme,
  topic,
  submitting,
  error,
  onSubmit,
}: {
  theme: TipsTheme;
  topic: TipTopic;
  submitting: boolean;
  error?: boolean;
  onSubmit: (text: string, hasImage: boolean) => void;
}) {
  const [text, setText] = useState('');
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const canSend = (text.trim().length > 0 || !!imagePreview) && !submitting;

  const handlePickImage = () => fileInputRef.current?.click();

  const handleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = () => setImagePreview(reader.result as string);
    reader.readAsDataURL(file);
    // Allow re-selecting the same file after a remove.
    e.target.value = '';
  };

  const handleSend = () => {
    if (!canSend) return;
    onSubmit(text.trim(), !!imagePreview);
  };

  return (
    <div style={{ padding: '4px 16px 32px' }}>
      <h1
        style={{
          margin: '0 0 12px',
          fontFamily: FONT,
          fontWeight: 600,
          fontSize: 26,
          lineHeight: '30px',
          letterSpacing: '-0.26px',
          color: INK,
        }}
      >
        Write your tip
      </h1>

      <p
        style={{
          margin: '0 0 4px',
          fontFamily: FONT,
          fontWeight: 600,
          fontSize: 20,
          lineHeight: '26px',
          letterSpacing: '-0.1px',
          color: GREY[800],
        }}
      >
        {topic.title}
      </p>
      <p
        style={{
          margin: '0 0 20px',
          fontFamily: FONT,
          fontWeight: 400,
          fontSize: 15,
          lineHeight: '21px',
          color: GREY[600],
        }}
      >
        {topic.prompt}
      </p>

      {/* Input card */}
      <div
        style={{
          position: 'relative',
          minHeight: 220,
          boxSizing: 'border-box',
          padding: 16,
          paddingBottom: 64,
          borderRadius: 16,
          backgroundColor: GREY[0],
          border: `1px solid ${GREY[200]}`,
          boxShadow: '0px 6px 16px rgba(0,0,0,0.06)',
        }}
      >
        <textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder={topic.placeholder}
          rows={4}
          style={{
            width: '100%',
            border: 'none',
            outline: 'none',
            resize: 'none',
            background: 'transparent',
            fontFamily: FONT,
            fontWeight: 400,
            fontSize: 16,
            lineHeight: '22px',
            color: GREY[800],
            boxSizing: 'border-box',
          }}
        />

        {/* Image thumbnail (bottom-left) with remove button */}
        {imagePreview && (
          <div
            style={{
              position: 'absolute',
              left: 16,
              bottom: 16,
              width: 76,
              height: 76,
              borderRadius: 12,
              overflow: 'hidden',
            }}
          >
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={imagePreview}
              alt="Attached"
              style={{ width: '100%', height: '100%', objectFit: 'cover' }}
            />
            <button
              type="button"
              onClick={() => setImagePreview(null)}
              aria-label="Remove image"
              style={{
                position: 'absolute',
                top: '50%',
                left: '50%',
                transform: 'translate(-50%, -50%)',
                width: 28,
                height: 28,
                borderRadius: '50%',
                border: 'none',
                cursor: 'pointer',
                backgroundColor: 'rgba(38,38,38,0.6)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                padding: 0,
              }}
            >
              <svg
                width="14"
                height="14"
                viewBox="0 0 24 24"
                fill="none"
                stroke="#ffffff"
                strokeWidth="2.5"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <line x1="18" y1="6" x2="6" y2="18" />
                <line x1="6" y1="6" x2="18" y2="18" />
              </svg>
            </button>
          </div>
        )}

        {/* Action buttons (bottom-right): attach image + send */}
        <div
          style={{
            position: 'absolute',
            right: 16,
            bottom: 16,
            display: 'flex',
            alignItems: 'center',
            gap: 12,
          }}
        >
          <button
            type="button"
            onClick={handlePickImage}
            aria-label="Attach image"
            style={{
              width: 44,
              height: 44,
              borderRadius: 10,
              border: 'none',
              cursor: 'pointer',
              backgroundColor: theme.imageBtnBg,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              padding: 0,
            }}
          >
            <Images size={22} color={theme.imageBtnFg} strokeWidth={1.75} />
          </button>

          <button
            type="button"
            onClick={handleSend}
            disabled={!canSend}
            aria-label="Send tip"
            style={{
              width: 44,
              height: 44,
              borderRadius: 10,
              border: 'none',
              cursor: canSend ? 'pointer' : 'default',
              backgroundColor: canSend ? GREY[800] : GREY[300],
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              padding: 0,
              transition: 'background-color 150ms ease',
            }}
          >
            <Send size={20} color="#ffffff" strokeWidth={2} />
          </button>
        </div>

        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          onChange={handleImageChange}
          style={{ display: 'none' }}
        />
      </div>

      {error && (
        <p
          role="alert"
          style={{
            margin: '12px 0 0',
            fontFamily: FONT,
            fontWeight: 500,
            fontSize: 14,
            lineHeight: '18px',
            color: '#D92D20',
          }}
        >
          Couldn’t share your tip. Please try again.
        </p>
      )}

      {/* Suggestion chips */}
      <div
        style={{
          marginTop: 16,
          display: 'flex',
          flexWrap: 'wrap',
          gap: 12,
        }}
      >
        {topic.suggestions.map((s) => {
          const active = text.trim().toLowerCase() === s.toLowerCase();
          return (
            <button
              key={s}
              type="button"
              onClick={() => setText(s)}
              style={{
                padding: '10px 18px',
                borderRadius: 999,
                cursor: 'pointer',
                fontFamily: FONT,
                fontWeight: 500,
                fontSize: 14,
                lineHeight: '18px',
                color: GREY[700],
                backgroundColor: active ? GREY[100] : GREY[0],
                border: `1px solid ${active ? GREY[400] : GREY[300]}`,
                transition: 'background-color 150ms ease, border-color 150ms ease',
              }}
            >
              {s}
            </button>
          );
        })}
      </div>
    </div>
  );
}
