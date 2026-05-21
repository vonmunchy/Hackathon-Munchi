interface ValidationResult {
  valid: boolean;
  error?: string;
}

export function validateName(name: string): ValidationResult {
  const trimmed = name.trim();
  if (trimmed.length < 2) {
    return { valid: false, error: "Name must be at least 2 characters" };
  }
  return { valid: true };
}

export function validatePhone(phone: string): ValidationResult {
  const digits = phone.replace(/\D/g, "");
  if (digits.length !== 7) {
    return { valid: false, error: "Phone number must be 7 digits" };
  }
  if (!/^[79]/.test(digits)) {
    return { valid: false, error: "Phone must start with 7 or 9" };
  }
  return { valid: true };
}

export function validateAddress(address: string): ValidationResult {
  const trimmed = address.trim();
  if (trimmed.length < 5) {
    return { valid: false, error: "Address must be at least 5 characters" };
  }
  return { valid: true };
}
