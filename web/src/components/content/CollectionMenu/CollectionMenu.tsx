import { BookmarkAction } from "src/components/site/Action/Bookmark";

import { Props, useToggleSave } from "./useCollectionMenu";

export function CollectionMenu(props: Props) {
  const { isSaved, onToggle } = useToggleSave(props);

  return (
    <BookmarkAction
      variant="subtle"
      size="xs"
      bookmarked={isSaved}
      onClick={onToggle}
    />
  );
}
