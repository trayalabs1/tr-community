import { useSession } from "@/auth";

export interface Props {
  initialSession: any;
}

export function useInfoScreen(props: Props) {
  const session = useSession(props.initialSession);

  return {
    session,
  };
}
