"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Button } from "@/components/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/dialog";
import { Form, FormField, SubmitButton } from "@/components/form";
import { Input } from "@/components/input";
import { getSwantaraService } from "@/lib/services/swantara";

type PlanningRunDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  onSave: () => void;
};

function usePlanningRunFormSchema() {
  return z.object({
    name: z.string(),
  });
}

export function PlanningRunDialog({ open, onOpenChange, orgId, onSave }: PlanningRunDialogProps) {
  const schema = usePlanningRunFormSchema();
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
    },
  });

  function handleSubmit() {
    void getSwantaraService()
      .planning.runs.create(Number(orgId), {
        horizonDays: 30,
      })
      .then(() => {
        toast.success("Planning run started");
        onSave();
      });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{"Run Planning"}</DialogTitle>
          <DialogDescription>
            {"Create a new Planning run to calculate material requirements"}
          </DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <FormField name="name" label={"Run Name"}>
            {({ field, id }) => <Input {...field} id={id} placeholder={"Optional name"} />}
          </FormField>
          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {"Cancel"}
            </Button>
            <SubmitButton>{"Run Planning"}</SubmitButton>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
