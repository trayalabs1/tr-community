import type { Metadata } from "next";

import { collectionList } from "@/api/openapi-server/collections";
import { UnreadyBanner } from "@/components/site/Unready";
import { SavedPostsScreen } from "@/screens/collection/SavedPostsScreen";

export const metadata: Metadata = {
  title: "Saved Posts",
};

export default async function Page() {
  try {
    const { data } = await collectionList();

    return <SavedPostsScreen initialCollections={data} />;
  } catch (e) {
    return <UnreadyBanner error={e} />;
  }
}
