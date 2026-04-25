export const STATUS_STYLES: Record<string, string> = {
  OK:           'text-primary border-primary/30 bg-primary/8',
  GUARDRAIL:    'text-amber-400 border-amber-400/30 bg-amber-400/8',
  'PII MASKED': 'text-violet-400 border-violet-400/30 bg-violet-400/8',
  FALLBACK:     'text-amber-400 border-amber-400/30 bg-amber-400/8',
  ERROR:        'text-red-400 border-red-400/30 bg-red-400/8',
}

export const ACTION_STYLES: Record<string, string> = {
  block:   'text-red-400 border-red-400/30 bg-red-400/8',
  rewrite: 'text-amber-400 border-amber-400/30 bg-amber-400/8',
  tag:     'text-primary border-primary/30 bg-primary/8',
  log:     'text-muted-foreground border-border bg-muted/40',
  shadow:  'text-violet-400 border-violet-400/30 bg-violet-400/8',
}

export const CIRCUIT_STYLES: Record<string, string> = {
  CLOSED:      'text-green-400 border-green-400/30 bg-green-400/8',
  OPEN:        'text-red-400 border-red-400/30 bg-red-400/8',
  'HALF-OPEN': 'text-amber-400 border-amber-400/30 bg-amber-400/8',
}

export const PII_TYPE_STYLES: Record<string, string> = {
  EMAIL:         'text-primary border-primary/30 bg-primary/8',
  PHONE:         'text-violet-400 border-violet-400/30 bg-violet-400/8',
  FULL_NAME:     'text-green-400 border-green-400/30 bg-green-400/8',
  SSN:           'text-red-400 border-red-400/30 bg-red-400/8',
  BANK_ACCT:     'text-amber-400 border-amber-400/30 bg-amber-400/8',
  ADDRESS:       'text-blue-400 border-blue-400/30 bg-blue-400/8',
  DATE_OF_BIRTH: 'text-pink-400 border-pink-400/30 bg-pink-400/8',
}

export function scoreColor(score: number): string {
  return score > 0.8 ? 'text-green-400' : score > 0.6 ? 'text-amber-400' : 'text-red-400'
}

export function quotaColor(pct: number): string {
  return pct > 80 ? 'text-red-400' : pct > 60 ? 'text-amber-400' : 'text-green-400'
}
