import { LikeSavedIcon } from "@/components/ui/icons/Like";
import { styled } from "@/styled-system/jsx";

type Props = {
  joined: string;
  likeScore: number;
};

export function ProfileStats({ joined, likeScore }: Props) {
  const memberSince = new Date(joined).toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
  });

  return (
    <styled.div
      display="flex"
      alignItems="stretch"
      w="full"
      py="4"
      borderRadius="2xl"
      bg="bg.profileStats"
    >
      <StatCell label="Member since" value={memberSince} />

      <styled.div w="[1px]" flexShrink="0" bg="border.default" />

      <StatCell
        label="Likes received"
        value={likeScore}
        icon={<LikeSavedIcon width="4" height="4" color="traya.accent.heart" />}
      />
    </styled.div>
  );
}

function StatCell({
  label,
  value,
  icon,
}: {
  label: string;
  value: string | number;
  icon?: React.ReactNode;
}) {
  return (
    <styled.div
      display="flex"
      flexDirection="column"
      alignItems="center"
      gap="1"
      flex="1"
      minW="0"
      px="3"
    >
      <styled.p fontSize="sm" lineHeight="[20px]" color="fg.muted">
        {label}
      </styled.p>

      <styled.div display="flex" alignItems="center" gap="1.5">
        {icon}
        <styled.p
          fontSize="md"
          lineHeight="[24px]"
          fontWeight="medium"
          color="fg.default"
        >
          {value}
        </styled.p>
      </styled.div>
    </styled.div>
  );
}
