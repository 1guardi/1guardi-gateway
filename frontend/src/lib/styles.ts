export const STATUS_STYLES: Record<string, string> = {
  OK:           'text-primary border-primary/30 bg-primary/8',
  GUARDRAIL:    'text-warning border-warning/30 bg-warning/8',
  'PII MASKED': 'text-violet border-violet/30 bg-violet/8',
  FALLBACK:     'text-warning border-warning/30 bg-warning/8',
  ERROR:        'text-error border-error/30 bg-error/8',
}

export const ACTION_STYLES: Record<string, string> = {
  block:   'text-error border-error/30 bg-error/8',
  rewrite: 'text-warning border-warning/30 bg-warning/8',
  tag:     'text-primary border-primary/30 bg-primary/8',
  log:     'text-muted-foreground border-border bg-muted/40',
  shadow:  'text-violet border-violet/30 bg-violet/8',
}

export const CIRCUIT_STYLES: Record<string, string> = {
  CLOSED:      'text-success border-success/30 bg-success/8',
  OPEN:        'text-error border-error/30 bg-error/8',
  'HALF-OPEN': 'text-warning border-warning/30 bg-warning/8',
}

export const PII_TYPE_STYLES: Record<string, string> = {
  EMAIL:         'text-primary border-primary/30 bg-primary/8',
  PHONE:         'text-violet border-violet/30 bg-violet/8',
  FULL_NAME:     'text-success border-success/30 bg-success/8',
  SSN:           'text-error border-error/30 bg-error/8',
  BANK_ACCT:     'text-warning border-warning/30 bg-warning/8',
  ADDRESS:       'text-info border-info/30 bg-info/8',
  DATE_OF_BIRTH: 'text-pink border-pink/30 bg-pink/8',
}

export function scoreColor(score: number): string {
  return score > 0.8 ? 'text-success' : score > 0.6 ? 'text-warning' : 'text-error'
}

export function quotaColor(pct: number): string {
  return pct > 80 ? 'text-error' : pct > 60 ? 'text-warning' : 'text-success'
}
