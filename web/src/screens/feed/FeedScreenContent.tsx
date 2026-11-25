"use client";

import { useEffect } from "react";
import {
  CategoryList,
  CategoryListResult,
  NodeListResult,
  ThreadListResult,
} from "@/api/openapi-schema";
import { useSettingsContext } from "@/components/site/SettingsContext/SettingsContext";
import { useEventTracking } from "@/lib/moengage/useEventTracking";

import { CategoryIndexScreen } from "../category/CategoryIndexScreen";

import { LibraryFeedScreen } from "./LibraryFeedScreen/LibraryFeedScreen";
import { ThreadFeedScreen } from "./ThreadFeedScreen/ThreadFeedScreen";

export type InitialData = {
  threads?: ThreadListResult;
  page?: number;
  library?: NodeListResult;
  categories?: CategoryListResult;
};

type Props = {
  initialData: InitialData;
};

export function FeedScreenContent({ initialData }: Props) {
  const { feed } = useSettingsContext();
  const { trackLandedHome } = useEventTracking();

  useEffect(() => {
    trackLandedHome();
  }, [trackLandedHome]);

  switch (feed.source.type) {
    case "threads":
      return (
        <ThreadFeedScreen
          initialPage={initialData.page}
          initialPageData={initialData.threads}
          category={undefined}
          paginationBasePath="/"
          showCategorySelect={true}
          showQuickShare={feed.source.quickShare === "enabled"}
        />
      );

    case "library":
      return <LibraryFeedScreen initialData={initialData.library} />;

    case "categories":
      return (
        <CategoryIndexScreen
          layout={feed.layout.type}
          threadListMode={feed.source.threadListMode}
          showQuickShare={feed.source.quickShare === "enabled"}
          initialCategoryList={initialData.categories}
          initialThreadList={initialData.threads}
          initialThreadListPage={initialData.page}
          paginationBasePath="/"
        />
      );
  }
}
