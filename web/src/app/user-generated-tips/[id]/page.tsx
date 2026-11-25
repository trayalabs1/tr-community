import { Suspense } from "react";
import type { Metadata } from "next";

import { getServerSession } from "@/auth/server-session";
import UserGeneratedTipsClient from "@/screens/user-generated-tips/UserGeneratedTipsClient";
import type { Gender } from "@/screens/user-generated-tips/tipsData";

export const metadata: Metadata = {
  title: "Traya | Share a Tip",
  description: "Share a tip to help someone in their first month",
};

type Props = {
  params: Promise<{ id: string }>;
  searchParams: Promise<{ gender?: string }>;
};

export default async function Page(props: Props) {
  const { id: caseId } = await props.params;
  const { gender: genderParam } = await props.searchParams;

  let gender: Gender = "M";
  const forced = (genderParam || "").toUpperCase();
  if (forced === "M" || forced === "F") gender = forced;

  let firstName: string | null = null;
  try {
    const session = await getServerSession();
    firstName = session?.name || session?.handle || null;
  } catch (error) {
    console.error("user-generated-tips: failed to load session", error);
  }

  return (
    <Suspense fallback={null}>
      <UserGeneratedTipsClient
        caseId={caseId}
        gender={gender}
        firstName={firstName}
      />
    </Suspense>
  );
}
