"use client";

import { useState, useMemo, useEffect, useCallback } from "react";
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

const MAX_CUSTOM_RANGE_DAYS = 3;

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

  const [pendingRangeStart, setPendingRangeStart] = useState<DateValue | null>(null);

  // While only the start of a range is picked, disable days that would make the range exceed the max.
  const isDateUnavailable = useCallback(
    (date: DateValue) =>
      pendingRangeStart !== null &&
      Math.abs(date.compare(pendingRangeStart)) > MAX_CUSTOM_RANGE_DAYS - 1,
    [pendingRangeStart],
  );

  const handleOpenChange = useCallback(({ open }: { open: boolean }) => {
    if (!open) setPendingRangeStart(null);
  }, []);

  const handleDateChange = ({ value }: { value: DateValue[] }) => {
    const [start, end] = value;
    setPendingRangeStart(start && !end ? start : null);
    if (!start || !end) return;

    const tz = getLocalTimeZone();
    const [earlier, later] = start.compare(end) <= 0 ? [start, end] : [end, start];
    const lastAllowed = earlier.add({ days: MAX_CUSTOM_RANGE_DAYS - 1 });
    const clampedLater = later.compare(lastAllowed) > 0 ? lastAllowed : later;
    setCustomDates({ start: earlier.toDate(tz), end: clampedLater.toDate(tz) });
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
    isDateUnavailable,
    handleOpenChange,
    handleApplyCustom,
    appliedCustomRange,
    data,
    isLoading,
  };
}
