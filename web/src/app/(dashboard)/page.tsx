import { redirect } from "next/navigation";
import { getServerSession } from "@/auth/server-session";

export default async function Page() {
  const session = await getServerSession();
  if (session) {
    redirect("/channels");
  }
}
