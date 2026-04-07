"use client";

import { useState, useMemo, useEffect } from "react";
import { today, getLocalTimeZone, type DateValue } from "@internationalized/date";

import { useAdminAnalyticsGet } from "@/api/openapi-client/admin";

function useDebounce<T>(value: T, delay: number): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const id = setTimeout(() => {
      setDebounced((prev) => (prev === value ? prev : value));
    }, delay);
    return () => clearTimeout(id);
  }, [value, delay]);
  return debounced;
}

export type FilterMode = "today" | "custom";

function todayAtHour(hour: number): string {
  const d = new Date();
  d.setHours(hour, 0, 0, 0);
  return d.toISOString();
}

export function useAnalyticsScreen() {
  const todayVal = useMemo(() => today(getLocalTimeZone()), []);

  const [mode, setMode] = useState<FilterMode>("today");
  const [startHour, setStartHour] = useState(0);
  const [endHour, setEndHour] = useState(23);
  const [customStartHour, setCustomStartHour] = useState(0);
  const [customEndHour, setCustomEndHour] = useState(23);
  const [customDates, setCustomDates] = useState<{ start: Date; end: Date }>(() => {
    const tz = getLocalTimeZone();
    const t = todayVal.toDate(tz);
    return { start: new Date(t), end: new Date(t) };
  });
  const [appliedCustomRange, setAppliedCustomRange] = useState<{ start: string; end: string } | null>(null);

  const handleDateChange = ({ value }: { value: DateValue[] }) => {
    const [start, end] = value;
    if (!start || !end) return;

    const tz = getLocalTimeZone();
    const [earlier, later] = start.compare(end) <= 0 ? [start, end] : [end, start];
    setCustomDates({ start: earlier.toDate(tz), end: later.toDate(tz) });
  };

  const handleApplyCustom = () => {
    const s = new Date(customDates.start);
    s.setHours(customStartHour, 0, 0, 0);
    const e = new Date(customDates.end);
    e.setHours(customEndHour, 59, 59, 999);
    setAppliedCustomRange({ start: s.toISOString(), end: e.toISOString() });
  };

  const start = mode === "today" ? todayAtHour(startHour) : (appliedCustomRange?.start ?? "");
  const end = mode === "today" ? todayAtHour(endHour) : (appliedCustomRange?.end ?? "");

  const debouncedStart = useDebounce(start, 600);
  const debouncedEnd = useDebounce(end, 600);

  const hasValidRange = debouncedStart !== "" && debouncedEnd !== "";
  const { data, isLoading } = useAdminAnalyticsGet(
    { start: debouncedStart, end: debouncedEnd },
    { swr: { enabled: hasValidRange } },
  );

  return {
    todayVal,
    mode,
    setMode,
    startHour,
    setStartHour,
    endHour,
    setEndHour,
    customStartHour,
    setCustomStartHour,
    customEndHour,
    setCustomEndHour,
    handleDateChange,
    handleApplyCustom,
    appliedCustomRange,
    data,
    isLoading,
  };
}
