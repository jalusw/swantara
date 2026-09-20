import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import {
  Questionnaire,
  QuestionnaireActions,
  QuestionnaireChoice,
  QuestionnaireChoiceInput,
  QuestionnaireChoices,
  QuestionnaireError,
  QuestionnaireItem,
  QuestionnaireNext,
  QuestionnairePrevious,
  QuestionnaireProgress,
  QuestionnaireSubmit,
} from "../questionnaire";

const items = [
  { name: "company", required: true },
  { name: "size", required: true },
];

function renderQuestionnaire(onSubmit: () => void = () => {}) {
  return renderWithProviders(
    <Questionnaire items={items} onSubmit={onSubmit}>
      <QuestionnaireProgress />
      <QuestionnaireItem name="company" required>
        <QuestionnaireChoices>
          <QuestionnaireChoice value="acme">
            <QuestionnaireChoiceInput />
            Acme
          </QuestionnaireChoice>
        </QuestionnaireChoices>
        <QuestionnaireError>Company is required</QuestionnaireError>
      </QuestionnaireItem>
      <QuestionnaireItem name="size" required>
        <QuestionnaireChoices>
          <QuestionnaireChoice value="small">
            <QuestionnaireChoiceInput />
            Small
          </QuestionnaireChoice>
        </QuestionnaireChoices>
        <QuestionnaireError>Size is required</QuestionnaireError>
      </QuestionnaireItem>
      <QuestionnaireActions>
        <QuestionnairePrevious>Back</QuestionnairePrevious>
        <QuestionnaireNext>Next</QuestionnaireNext>
        <QuestionnaireSubmit>Submit</QuestionnaireSubmit>
      </QuestionnaireActions>
    </Questionnaire>,
  );
}

describe("Questionnaire", () => {
  it("shows progress and only the first item", () => {
    renderQuestionnaire();

    expect(screen.getByRole("progressbar")).toBeInTheDocument();
    expect(screen.getByText("Acme")).toBeVisible();
    expect(screen.getByText("Small")).not.toBeVisible();
  });

  it("blocks next on unanswered required item and advances after answering", async () => {
    const user = userEvent.setup();
    renderQuestionnaire();

    await user.click(screen.getByRole("button", { name: "Next" }));
    expect(screen.getByText("Small")).not.toBeVisible();

    await user.click(screen.getByText("Acme"));
    await user.click(screen.getByRole("button", { name: "Next" }));
    expect(screen.getByText("Small")).toBeVisible();
  });

  it("navigates back and submits on the last item", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    renderQuestionnaire(onSubmit);

    expect(screen.queryByRole("button", { name: "Submit" })).not.toBeInTheDocument();

    await user.click(screen.getByText("Acme"));
    await user.click(screen.getByRole("button", { name: "Next" }));
    await user.click(screen.getByText("Small"));
    await user.click(screen.getByRole("button", { name: "Submit" }));

    expect(onSubmit).toHaveBeenCalled();
  });
});
