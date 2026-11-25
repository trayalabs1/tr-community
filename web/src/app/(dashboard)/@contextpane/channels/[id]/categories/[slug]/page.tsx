import {
  channelCategoryGet,
  channelThreadList,
} from "@/api/openapi-server/channels";
import { CategoryScreenContextPane } from "@/screens/category/CategoryScreenContextPane";

export default async function Page(props: {
  params: Promise<{ id: string; slug: string }>;
}) {
  const { id: channelID, slug } = await props.params;

  try {
    const { data: categoryData } = await channelCategoryGet(channelID, slug);

    const { data: threadListData } = await channelThreadList(channelID, {
      categories: [slug],
    });

    return (
      <CategoryScreenContextPane
        slug={slug}
        initialCategory={categoryData}
        initialThreadList={threadListData}
      />
    );
  } catch (e) {
    console.error(e);
    return null;
  }
}
