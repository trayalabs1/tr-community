import { Suspense } from "react";
import { UnreadyBanner } from "@/components/site/Unready";
import { QueueScreen } from "@/screens/queue/QueueScreen";

export default async function Page() {
  try {
    return (
      <Suspense>
        <QueueScreen />
      </Suspense>
    );
  } catch (e) {
    return <UnreadyBanner error={e} />;
  }
}
