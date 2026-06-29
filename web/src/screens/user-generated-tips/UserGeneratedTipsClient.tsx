'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import SubpageTopBar from './SubpageTopBar';
import { moengageTrackEvent } from './analytics';
import TopicSelection from './TopicSelection';
import WriteTip from './WriteTip';
import TipSubmitted from './TipSubmitted';
import { getTheme, getTopic, GREY, type Gender } from './tipsData';
import { submitTip } from './tipsApi';
import './tokens.css';

type Step = 'topic' | 'write' | 'submitted';

const COMPONENT = 'user_tips';

/** Exit the whole flow — hand back to the React Native shell when embedded. */
function closeFlow() {
  if (typeof window === 'undefined') return;
  const w = window as any;
  if (w.__rnGoBack) w.__rnGoBack();
  else if (w.ReactNativeWebView) w.ReactNativeWebView.postMessage('goBack');
  else window.history.back();
}

export default function UserGeneratedTipsClient({
  caseId,
  gender,
  firstName,
  fontClassName,
}: {
  caseId: string;
  gender: Gender;
  firstName?: string | null;
  fontClassName?: string;
}) {
  const [step, setStep] = useState<Step>('topic');
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const theme = useMemo(() => getTheme(gender), [gender]);
  const topic = getTopic(selectedId);
  const advanceTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  // User lands on the web page.
  useEffect(() => {
    moengageTrackEvent('web_component_item_viewed', {
      component: COMPONENT,
      case_id: caseId,
    });
    return () => {
      if (advanceTimer.current) clearTimeout(advanceTimer.current);
    };
  }, [caseId]);

  const handleSelectTopic = (id: string) => {
    setSelectedId(id);
    // User selects a topic. Distinct from the page-view event above so it does
    // not double-count the "viewed" funnel step (was web_component_item_viewed).
    moengageTrackEvent('web_component_item_clicked', {
      component: COMPONENT,
      topic: getTopic(id)?.title ?? id,
      case_id: caseId,
    });
    // Let the green selected state render briefly before moving to the
    // write step (the topic screen has no explicit Continue button).
    if (advanceTimer.current) clearTimeout(advanceTimer.current);
    advanceTimer.current = setTimeout(() => setStep('write'), 300);
  };

  const handleSubmit = async (text: string, hasImage: boolean) => {
    if (!topic || submitting) return;
    // User clicks submit.
    moengageTrackEvent('app_component_item_clicked', {
      component: COMPONENT,
      topic: topic.title,
      case_id: caseId,
    });
    setSubmitting(true);
    // Advance to the confirmation regardless of network outcome — the tips
    // endpoint isn't live yet (see tipsApi.ts). Errors are logged there.
    await submitTip({
      caseId,
      topicId: topic.id,
      topicTitle: topic.title,
      text,
      hasImage,
    });
    setSubmitting(false);
    setStep('submitted');
  };

  const handleShareAnother = () => {
    setSelectedId(null);
    setStep('topic');
  };

  // Back from the write step returns to a clean topic list — clear the previous
  // selection so no card stays highlighted, and cancel any pending auto-advance.
  const handleBackToTopics = () => {
    if (advanceTimer.current) clearTimeout(advanceTimer.current);
    setSelectedId(null);
    setStep('topic');
  };

  return (
    <main
      className={`ugt-scope ${fontClassName ?? ''}`}
      style={{
        marginLeft: 'auto',
        marginRight: 'auto',
        display: 'flex',
        width: '100%',
        maxWidth: '28rem',
        flexDirection: 'column',
        minHeight: '100vh',
        backgroundColor: GREY[0],
        background: theme.pageGradient,
        backgroundRepeat: 'no-repeat',
      }}
    >
      {/* Transparent top bar: the page gradient flows behind it with no solid
          band (mirrors Figma's transparent app bar). */}
      {step === 'topic' && (
        <SubpageTopBar onBack={closeFlow} backgroundColor="transparent" />
      )}
      {step === 'write' && (
        <SubpageTopBar
          onBack={handleBackToTopics}
          backgroundColor="transparent"
        />
      )}
      {step === 'submitted' && (
        <SubpageTopBar hideBack onClose={closeFlow} backgroundColor="transparent" />
      )}

      {step === 'topic' && (
        <TopicSelection
          theme={theme}
          selectedId={selectedId}
          onSelect={handleSelectTopic}
        />
      )}

      {step === 'write' && topic && (
        <WriteTip
          theme={theme}
          topic={topic}
          submitting={submitting}
          onSubmit={handleSubmit}
        />
      )}

      {step === 'submitted' && (
        <TipSubmitted
          theme={theme}
          name={firstName}
          onGoHome={closeFlow}
          onShareAnother={handleShareAnother}
        />
      )}
    </main>
  );
}
