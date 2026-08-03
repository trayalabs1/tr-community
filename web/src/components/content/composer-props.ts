import { Asset } from "@/api/openapi-schema";

export type ContentComposerProps = {
  className?: string;
  disabled?: boolean;
  resetKey?: string;
  initialValue?: string;

  // NOTE: This is not for making the editor controllable but for optimistic
  // mutation/revalidation of disabled editors. Use with care!
  value?: string;
  placeholder?: string;
  onChange?: (value: string, isEmpty: boolean) => void;
  onAssetUpload?: (asset: Asset) => void;

  // Suppresses the floating ComposerTools toolbar for surfaces that provide
  // their own controls.
  hideTools?: boolean;
};
