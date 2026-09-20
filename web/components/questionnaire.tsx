import { Questionnaire as QuestionnairePrimitive } from "@shadcn/react/questionnaire";
import type * as React from "react";
import { buttonVariants } from "@/components/button";
import { cn } from "@/lib/utils";

type QuestionnaireProps = React.ComponentProps<typeof QuestionnairePrimitive.Root>;

function Questionnaire({ className, ...props }: QuestionnaireProps) {
  return (
    <QuestionnairePrimitive.Root
      data-slot="questionnaire"
      className={cn("flex w-full flex-col gap-y-6", className)}
      {...props}
    />
  );
}

type QuestionnaireProgressProps = React.ComponentProps<typeof QuestionnairePrimitive.Progress>;

function QuestionnaireProgress({ className, ...props }: QuestionnaireProgressProps) {
  return (
    <QuestionnairePrimitive.Progress
      data-slot="questionnaire-progress"
      render={(elementProps, state) => (
        <div
          {...(elementProps as React.ComponentProps<"div">)}
          className={cn("h-1.5 w-full overflow-hidden rounded-full bg-muted", className)}
        >
          <div
            className="h-full rounded-full bg-primary transition-[width]"
            style={{ width: state.total > 0 ? `${(state.current / state.total) * 100}%` : "0%" }}
          />
        </div>
      )}
      {...props}
    />
  );
}

type QuestionnaireItemProps = React.ComponentProps<typeof QuestionnairePrimitive.Item>;

function QuestionnaireItem({ className, ...props }: QuestionnaireItemProps) {
  return (
    <QuestionnairePrimitive.Item
      data-slot="questionnaire-item"
      className={cn("flex flex-col gap-y-4", className)}
      {...props}
    />
  );
}

type QuestionnaireTitleProps = React.ComponentProps<typeof QuestionnairePrimitive.Title>;

function QuestionnaireTitle({ className, ...props }: QuestionnaireTitleProps) {
  return (
    <QuestionnairePrimitive.Title
      data-slot="questionnaire-title"
      className={cn("text-lg leading-none", className)}
      {...props}
    />
  );
}

type QuestionnaireDescriptionProps = React.ComponentProps<
  typeof QuestionnairePrimitive.Description
>;

function QuestionnaireDescription({ className, ...props }: QuestionnaireDescriptionProps) {
  return (
    <QuestionnairePrimitive.Description
      data-slot="questionnaire-description"
      className={cn("mt-2 text-sm leading-relaxed text-muted-foreground", className)}
      {...props}
    />
  );
}

type QuestionnaireChoicesProps = React.ComponentProps<typeof QuestionnairePrimitive.Choices>;

function QuestionnaireChoices({ className, ...props }: QuestionnaireChoicesProps) {
  return (
    <QuestionnairePrimitive.Choices
      data-slot="questionnaire-choices"
      className={cn("flex flex-col gap-y-2", className)}
      {...props}
    />
  );
}

type QuestionnaireChoiceProps = React.ComponentProps<typeof QuestionnairePrimitive.Choice>;

function QuestionnaireChoice({ className, ...props }: QuestionnaireChoiceProps) {
  return (
    <QuestionnairePrimitive.Choice
      data-slot="questionnaire-choice"
      className={cn(
        "flex min-h-[44px] cursor-pointer items-center gap-3 rounded-lg border border-border px-4 py-3 text-sm transition-colors has-checked:border-primary has-checked:bg-primary/5 has-disabled:cursor-not-allowed has-disabled:opacity-50",
        className,
      )}
      {...props}
    />
  );
}

type QuestionnaireChoiceInputProps = React.ComponentProps<
  typeof QuestionnairePrimitive.ChoiceInput
>;

function QuestionnaireChoiceInput({ className, ...props }: QuestionnaireChoiceInputProps) {
  return (
    <QuestionnairePrimitive.ChoiceInput
      data-slot="questionnaire-choice-input"
      className={cn("size-4 shrink-0 accent-primary", className)}
      {...props}
    />
  );
}

type QuestionnaireInputProps = React.ComponentProps<typeof QuestionnairePrimitive.Input>;

function QuestionnaireInput({ className, ...props }: QuestionnaireInputProps) {
  return (
    <QuestionnairePrimitive.Input
      data-slot="questionnaire-input"
      className={cn(
        "min-h-[44px] w-full rounded-lg border border-input bg-transparent bg-clip-padding px-3 py-2 text-sm transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 dark:bg-input/30",
        className,
      )}
      {...props}
    />
  );
}

type QuestionnaireErrorProps = React.ComponentProps<typeof QuestionnairePrimitive.Error>;

function QuestionnaireError({ className, ...props }: QuestionnaireErrorProps) {
  return (
    <QuestionnairePrimitive.Error
      data-slot="questionnaire-error"
      className={cn("text-sm text-destructive", className)}
      {...props}
    />
  );
}

type QuestionnaireActionsProps = React.ComponentProps<"div">;

function QuestionnaireActions({ className, ...props }: QuestionnaireActionsProps) {
  return (
    <div
      data-slot="questionnaire-actions"
      className={cn("mt-2 flex items-center justify-between gap-2", className)}
      {...props}
    />
  );
}

type QuestionnairePreviousProps = React.ComponentProps<typeof QuestionnairePrimitive.Previous>;

function QuestionnairePrevious({ className, ...props }: QuestionnairePreviousProps) {
  return (
    <QuestionnairePrimitive.Previous
      data-slot="questionnaire-previous"
      className={cn(buttonVariants({ variant: "outline", size: "lg" }), className)}
      {...props}
    />
  );
}

type QuestionnaireNextProps = React.ComponentProps<typeof QuestionnairePrimitive.Next>;

function QuestionnaireNext({ className, ...props }: QuestionnaireNextProps) {
  return (
    <QuestionnairePrimitive.Next
      data-slot="questionnaire-next"
      className={cn(buttonVariants({ variant: "default", size: "lg" }), className)}
      {...props}
    />
  );
}

type QuestionnaireSubmitProps = React.ComponentProps<typeof QuestionnairePrimitive.Submit>;

function QuestionnaireSubmit({ className, ...props }: QuestionnaireSubmitProps) {
  return (
    <QuestionnairePrimitive.Submit
      data-slot="questionnaire-submit"
      className={cn(buttonVariants({ variant: "default", size: "lg" }), className)}
      {...props}
    />
  );
}

export {
  Questionnaire,
  QuestionnaireActions,
  QuestionnaireChoice,
  QuestionnaireChoiceInput,
  QuestionnaireChoices,
  QuestionnaireDescription,
  QuestionnaireError,
  QuestionnaireInput,
  QuestionnaireItem,
  QuestionnaireNext,
  QuestionnairePrevious,
  QuestionnaireProgress,
  QuestionnaireSubmit,
  QuestionnaireTitle,
};
