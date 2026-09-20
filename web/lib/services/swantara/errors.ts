export type SwantaraFieldError = {
  field: string;
  code: string;
  message: string;
};

export class SwantaraError extends Error {
  readonly code: string | null;
  readonly status: number;

  constructor(message: string, code: string | null = null, status = 0) {
    super(message);
    this.name = "SwantaraError";
    this.code = code;
    this.status = status;
  }
}

export class SwantaraUnauthorizedError extends SwantaraError {
  constructor(message: string) {
    super(message, "ERR_UNAUTHORIZED", 401);
  }
}

export class SwantaraUnprocessableError extends SwantaraError {
  fieldErrors: SwantaraFieldError[] | null;

  constructor(message: string, fieldErrors: SwantaraFieldError[] | null = null) {
    super(message, "ERR_UNPROCESSABLE", 422);
    this.fieldErrors = fieldErrors;
  }
}
