export function equipmentStateTone(
  state: string,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "active":
      return "success";
    case "inactive":
      return "neutral";
    case "maintenance":
      return "warning";
    case "retired":
      return "danger";
    default:
      return "neutral";
  }
}

export function formatEquipmentState(state: string): string {
  switch (state) {
    case "active":
      return "Active";
    case "inactive":
      return "Inactive";
    case "maintenance":
      return "Maintenance";
    case "retired":
      return "Retired";
    default:
      return state;
  }
}
