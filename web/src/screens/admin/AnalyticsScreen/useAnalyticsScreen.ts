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
  const [customRange, setCustomRange] = useState<{ start: string; end: string }>(() => {
    const tz = getLocalTimeZone();
    const t = todayVal.toDate(tz);
    const s = new Date(t);
    s.setHours(0, 0, 0, 0);
    const e = new Date(t);
    e.setHours(23, 59, 59, 999);
    return { start: s.toISOString(), end: e.toISOString() };
  });

  const handleDateChange = ({ value }: { value: DateValue[] }) => {
    const [start, end] = value;
    if (!start || !end) return;

    const tz = getLocalTimeZone();
    const [earlier, later] = start.compare(end) <= 0 ? [start, end] : [end, start];
    const startDate = earlier.toDate(tz);
    startDate.setHours(0, 0, 0, 0);
    const endDate = later.toDate(tz);
    endDate.setHours(23, 59, 59, 999);

    setCustomRange({ start: startDate.toISOString(), end: endDate.toISOString() });
  };

  const start = mode === "today" ? todayAtHour(startHour) : customRange.start;
  const end = mode === "today" ? todayAtHour(endHour) : customRange.end;

  const debouncedStart = useDebounce(start, 600);
  const debouncedEnd = useDebounce(end, 600);

  const { data, isLoading } = useAdminAnalyticsGet({ start: debouncedStart, end: debouncedEnd });

  return {
    todayVal,
    mode,
    setMode,
    startHour,
    setStartHour,
    endHour,
    setEndHour,
    handleDateChange,
    data,
    isLoading,
  };
}
