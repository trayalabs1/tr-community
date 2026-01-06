import { notFound } from "next/navigation";
import { LandingScreen } from "@/screens/auth/LandingScreen/LandingScreen";

type Props = {
  params: Promise<{
    token: string;
  }>;
};

export default async function Page(props: Props) {
  const { token } = await props.params;

  if (!token || token.trim() === "") {
    notFound();
  }

  return <LandingScreen token={token} />;
}
