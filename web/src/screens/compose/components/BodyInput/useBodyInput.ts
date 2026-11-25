import { useFormContext } from "react-hook-form";

import { FormShape } from "../ComposeForm/useComposeForm";

export function useBodyInput() {
  const ctx = useFormContext<FormShape>();

  // Only surface the error once the author has actually touched the body, so a
  // freshly opened composer does not open with a validation message.
  const isTouched =
    Boolean(ctx.formState.dirtyFields.body) || ctx.formState.isSubmitted;

  return {
    control: ctx.control,
    error: isTouched ? ctx.formState.errors.body?.message : undefined,
  };
}
