"use client";

import { css } from "@/styled-system/css";
import { CardBox, HStack, VStack } from "@/styled-system/jsx";
import { DateRangePicker } from "@/components/ui/date-picker";
import { useState, useEffect } from "react";

import { useAnalyticsScreen } from "./useAnalyticsScreen";

export function AnalyticsScreen() {
  const {
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
  } = useAnalyticsScreen();

  const [startInput, setStartInput] = useState(String(startHour));
  const [endInput, setEndInput] = useState(String(endHour));
  const [customStartInput, setCustomStartInput] = useState(String(customStartHour));
  const [customEndInput, setCustomEndInput] = useState(String(customEndHour));

  useEffect(() => { setStartInput(String(startHour)); }, [startHour]);
  useEffect(() => { setEndInput(String(endHour)); }, [endHour]);
  useEffect(() => { setCustomStartInput(String(customStartHour)); }, [customStartHour]);
  useEffect(() => { setCustomEndInput(String(customEndHour)); }, [customEndHour]);

  const totalOnboardings =
    data?.channel_onboardings?.reduce((s, r) => s + r.count, 0) ?? 0;
  const totalPosts =
    data?.channel_posts?.reduce((s, r) => s + r.count, 0) ?? 0;

  return (
    <VStack gap="6" alignItems="stretch" paddingTop="4">
      <HStack gap="3" alignItems="center" flexWrap="wrap">
        <HStack gap="0" style={{ border: "1px solid var(--colors-border-default)", borderRadius: "0.375rem", overflow: "hidden" }}>
          {(["today", "custom"] as const).map((m) => (
            <button
              key={m}
              type="button"
              onClick={() => setMode(m)}
              className={css({ px: "3", py: "1.5", fontSize: "sm", cursor: "pointer", fontWeight: mode === m ? "semibold" : "normal" })}
              style={{ background: mode === m ? "var(--colors-bg-muted)" : "transparent", border: "none", color: "var(--colors-fg-default)" }}
            >
              {m === "today" ? "Today" : "Custom"}
            </button>
          ))}
        </HStack>

        {mode === "today" && (
          <HStack gap="2" alignItems="center">
            <input
              type="number"
              min={0}
              max={23}
              value={startInput}
              onChange={(e) => {
                setStartInput(e.target.value);
                const n = parseInt(e.target.value, 10);
                if (!isNaN(n)) setStartHour(Math.min(23, Math.max(0, n)));
              }}
              onBlur={() => {
                const n = parseInt(startInput, 10);
                const clamped = isNaN(n) ? 0 : Math.min(23, Math.max(0, n));
                setStartHour(clamped);
                setStartInput(String(clamped));
              }}
              className={css({ width: "14", borderWidth: "thin", borderStyle: "solid", borderColor: "border.default", borderRadius: "md", px: "2", py: "1", fontSize: "sm" })}
            />
            <span className={css({ fontSize: "sm", color: "fg.muted" })}>to</span>
            <input
              type="number"
              min={0}
              max={23}
              value={endInput}
              onChange={(e) => {
                setEndInput(e.target.value);
                const n = parseInt(e.target.value, 10);
                if (!isNaN(n)) setEndHour(Math.min(23, Math.max(0, n)));
              }}
              onBlur={() => {
                const n = parseInt(endInput, 10);
                const clamped = isNaN(n) ? 23 : Math.min(23, Math.max(0, n));
                setEndHour(clamped);
                setEndInput(String(clamped));
              }}
              className={css({ width: "14", borderWidth: "thin", borderStyle: "solid", borderColor: "border.default", borderRadius: "md", px: "2", py: "1", fontSize: "sm" })}
            />
          </HStack>
        )}

        {mode === "custom" && (
          <HStack gap="2" alignItems="center" flexWrap="wrap">
            <DateRangePicker hideInputs={true} max={todayVal} onValueChange={handleDateChange} />
            <input
              type="number"
              min={0}
              max={23}
              value={customStartInput}
              onChange={(e) => {
                setCustomStartInput(e.target.value);
                const n = parseInt(e.target.value, 10);
                if (!isNaN(n)) setCustomStartHour(Math.min(23, Math.max(0, n)));
              }}
              onBlur={() => {
                const n = parseInt(customStartInput, 10);
                const clamped = isNaN(n) ? 0 : Math.min(23, Math.max(0, n));
                setCustomStartHour(clamped);
                setCustomStartInput(String(clamped));
              }}
              className={css({ width: "14", borderWidth: "thin", borderStyle: "solid", borderColor: "border.default", borderRadius: "md", px: "2", py: "1", fontSize: "sm" })}
            />
            <span className={css({ fontSize: "sm", color: "fg.muted" })}>to</span>
            <input
              type="number"
              min={0}
              max={23}
              value={customEndInput}
              onChange={(e) => {
                setCustomEndInput(e.target.value);
                const n = parseInt(e.target.value, 10);
                if (!isNaN(n)) setCustomEndHour(Math.min(23, Math.max(0, n)));
              }}
              onBlur={() => {
                const n = parseInt(customEndInput, 10);
                const clamped = isNaN(n) ? 23 : Math.min(23, Math.max(0, n));
                setCustomEndHour(clamped);
                setCustomEndInput(String(clamped));
              }}
              className={css({ width: "14", borderWidth: "thin", borderStyle: "solid", borderColor: "border.default", borderRadius: "md", px: "2", py: "1", fontSize: "sm" })}
            />
            <button
              type="button"
              onClick={handleApplyCustom}
              className={css({ px: "3", py: "1.5", fontSize: "sm", fontWeight: "semibold", cursor: "pointer", borderRadius: "md" })}
              style={{ background: "var(--colors-bg-muted)", border: "1px solid var(--colors-border-default)", color: "var(--colors-fg-default)" }}
            >
              Apply
            </button>
          </HStack>
        )}
      </HStack>

      {isLoading && <p>Loading…</p>}

      {mode === "custom" && !appliedCustomRange && !isLoading && (
        <p className={css({ fontSize: "sm", color: "fg.muted" })}>Select a date range and hours, then click Apply.</p>
      )}

      {data && (
        <HStack gap="4" alignItems="flex-start" flexWrap="wrap">
          <CardBox p="4" minWidth="[280px]">
            <ChannelTable
              title="Channel Distribution — Onboardings"
              colHeader="Users Onboarded"
              rows={data.channel_onboardings ?? []}
              total={totalOnboardings}
            />
          </CardBox>
          <CardBox p="4" minWidth="[280px]">
            <ChannelTable
              title="Channel Distribution — User Posts"
              colHeader="User Posts"
              rows={data.channel_posts ?? []}
              total={totalPosts}
            />
          </CardBox>
          <CardBox p="4" minWidth="[280px]">
            <AdminReplyTimeTable rows={data.admin_reply_times ?? []} />
          </CardBox>
        </HStack>
      )}
    </VStack>
  );
}

type ChannelRow = { channel_name: string; count: number };

function ChannelTable({
  title,
  colHeader,
  rows,
  total,
}: {
  title: string;
  colHeader: string;
  rows: ChannelRow[];
  total: number;
}) {
  return (
    <VStack gap="2" alignItems="stretch">
      <p className={css({ fontWeight: "semibold", fontSize: "sm" })}>{title}</p>
      <table
        className={css({
          width: "full",
          borderCollapse: "collapse",
          fontSize: "sm",
        })}
      >
        <thead>
          <tr>
            <th
              className={css({
                textAlign: "left",
                paddingX: "2",
                paddingY: "1",
                borderBottomWidth: "thin",
                borderBottomStyle: "solid",
                borderBottomColor: "border.default",
              })}
            >
              Channel
            </th>
            <th
              className={css({
                textAlign: "right",
                paddingX: "2",
                paddingY: "1",
                borderBottomWidth: "thin",
                borderBottomStyle: "solid",
                borderBottomColor: "border.default",
              })}
            >
              {colHeader}
            </th>
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={row.channel_name}>
              <td
                className={css({
                  paddingX: "2",
                  paddingY: "1",
                  borderBottomWidth: "thin",
                  borderBottomStyle: "solid",
                  borderBottomColor: "border.subtle",
                })}
              >
                {row.channel_name}
              </td>
              <td
                className={css({
                  paddingX: "2",
                  paddingY: "1",
                  textAlign: "right",
                  borderBottomWidth: "thin",
                  borderBottomStyle: "solid",
                  borderBottomColor: "border.subtle",
                })}
              >
                {row.count}
              </td>
            </tr>
          ))}
          <tr className={css({ fontWeight: "bold" })}>
            <td className={css({ paddingX: "2", paddingY: "1" })}>Total</td>
            <td className={css({ paddingX: "2", paddingY: "1", textAlign: "right" })}>
              {total}
            </td>
          </tr>
        </tbody>
      </table>
    </VStack>
  );
}

type ReplyRow = { admin_handle: string; avg_time_minutes: number };

function AdminReplyTimeTable({ rows }: { rows: ReplyRow[] }) {
  return (
    <VStack gap="2" alignItems="stretch">
      <p className={css({ fontWeight: "semibold", fontSize: "sm" })}>Admin Reply Time</p>
      <table
        className={css({
          width: "full",
          borderCollapse: "collapse",
          fontSize: "sm",
        })}
      >
        <thead>
          <tr>
            <th
              className={css({
                textAlign: "left",
                paddingX: "2",
                paddingY: "1",
                borderBottomWidth: "thin",
                borderBottomStyle: "solid",
                borderBottomColor: "border.default",
              })}
            >
              Admin Handle
            </th>
            <th
              className={css({
                textAlign: "right",
                paddingX: "2",
                paddingY: "1",
                borderBottomWidth: "thin",
                borderBottomStyle: "solid",
                borderBottomColor: "border.default",
              })}
            >
              Avg Time (mins)
            </th>
          </tr>
        </thead>
        <tbody>
          {rows.length === 0 ? (
            <tr>
              <td
                colSpan={2}
                className={css({
                  padding: "2",
                  textAlign: "center",
                  color: "fg.muted",
                })}
              >
                No data in range
              </td>
            </tr>
          ) : (
            rows.map((row) => (
              <tr key={row.admin_handle}>
                <td
                  className={css({
                    paddingX: "2",
                    paddingY: "1",
                    borderBottomWidth: "thin",
                    borderBottomStyle: "solid",
                    borderBottomColor: "border.subtle",
                  })}
                >
                  @{row.admin_handle}
                </td>
                <td
                  className={css({
                    paddingX: "2",
                    paddingY: "1",
                    textAlign: "right",
                    borderBottomWidth: "thin",
                    borderBottomStyle: "solid",
                    borderBottomColor: "border.subtle",
                  })}
                >
                  {row.avg_time_minutes.toFixed(1)}
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    </VStack>
  );
}
