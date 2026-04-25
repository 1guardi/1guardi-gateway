export interface Trace {
  id: string
  agent: string
  agentSub: string
  agentId?: string
  model: string
  inputTokens: number
  outputTokens: number
  cost: string
  ttft: string
  tps: string
  status: 'OK' | 'GUARDRAIL' | 'PII MASKED' | 'FALLBACK' | 'ERROR'
  ts: string
  threadId: string
}

export interface GuardrailRule {
  id: string
  priority: number
  name: string
  scope: string[]
  action: 'block' | 'rewrite' | 'tag' | 'log' | 'shadow'
  mode: 'parallel' | 'sequential'
  managed: boolean
  enabled: boolean
  fires24h: number
  agentId?: string
}

export interface PiiEntry {
  token: string
  type: string
  agent: string
  maskedAt: string
  ttl: string
  hits: number
}

export interface Endpoint {
  id: string
  label: string
  model: string
  region: string
  ttftP50: string
  ttftP99: string
  avgTps: string
  errorRate: string
  quotaUsed: number
  circuitState: 'CLOSED' | 'OPEN' | 'HALF-OPEN'
  score: number
}

export interface APIKey {
  id: string
  name: string
  prefix: string
  tenantId: string
  agentId?: string
  lastUsed: string
  isActive: boolean
  createdAt: string
}

export const traces: Trace[] = [
  { id: 'tr-001', agent: 'AGT-001', agentSub: 'customer-support', agentId: 'AGT-001', model: 'gpt-4o',     inputTokens: 812,  outputTokens: 435, cost: '$0.031', ttft: '89ms',  tps: '42', status: 'OK',        ts: '14:32:01', threadId: 'th-aaa1' },
  { id: 'tr-002', agent: 'AGT-002', agentSub: 'data-pipeline',    agentId: 'AGT-002', model: 'claude-3.5', inputTokens: 6120, outputTokens: 2310,cost: '$0.021', ttft: '234ms', tps: '38', status: 'GUARDRAIL',  ts: '14:31:47', threadId: 'th-bbb2' },
  { id: 'tr-003', agent: 'AGT-001', agentSub: 'customer-support', agentId: 'AGT-001', model: 'gpt-4o',     inputTokens: 1450, outputTokens: 653, cost: '$0.052', ttft: '112ms', tps: '39', status: 'PII MASKED', ts: '14:31:22', threadId: 'th-aaa1' },
  { id: 'tr-004', agent: 'AGT-003', agentSub: 'code-assistant',   model: 'gpt-4o',     inputTokens: 590,  outputTokens: 301, cost: '$0.022', ttft: '156ms', tps: '44', status: 'OK',        ts: '14:30:58', threadId: 'th-ccc3' },
  { id: 'tr-005', agent: 'AGT-002', agentSub: 'data-pipeline',    model: 'gemini-pro', inputTokens: 2800, outputTokens: 1012,cost: '$0.009', ttft: '310ms', tps: '29', status: 'FALLBACK',   ts: '14:30:33', threadId: 'th-bbb2' },
  { id: 'tr-006', agent: 'AGT-003', agentSub: 'code-assistant',   model: 'gpt-4o',     inputTokens: 720,  outputTokens: 289, cost: '$0.018', ttft: '99ms',  tps: '46', status: 'OK',        ts: '14:29:51', threadId: 'th-ccc3' },
  { id: 'tr-007', agent: 'AGT-001', agentSub: 'customer-support', model: 'gpt-4o',     inputTokens: 430,  outputTokens: 198, cost: '$0.015', ttft: '78ms',  tps: '51', status: 'GUARDRAIL',  ts: '14:29:10', threadId: 'th-ddd4' },
  { id: 'tr-008', agent: 'AGT-002', agentSub: 'data-pipeline',    model: 'claude-3.5', inputTokens: 5400, outputTokens: 1890,cost: '$0.019', ttft: '198ms', tps: '40', status: 'OK',        ts: '14:28:42', threadId: 'th-bbb2' },
  { id: 'tr-009', agent: 'AGT-001', agentSub: 'customer-support', model: 'gpt-4o',     inputTokens: 910,  outputTokens: 510, cost: '$0.035', ttft: '102ms', tps: '43', status: 'ERROR',     ts: '14:28:01', threadId: 'th-eee5' },
  { id: 'tr-010', agent: 'AGT-003', agentSub: 'code-assistant',   model: 'gpt-4o',     inputTokens: 640,  outputTokens: 310, cost: '$0.020', ttft: '134ms', tps: '41', status: 'PII MASKED', ts: '14:27:30', threadId: 'th-ccc3' },
]

export const guardrailRules: GuardrailRule[] = [
  { id: 'mgd-001', priority: 1,  name: 'Prompt Injection Detection',  scope: ['input'],            action: 'block',   mode: 'parallel',   managed: true,  enabled: true,  fires24h: 12 },
  { id: 'mgd-002', priority: 2,  name: 'PII Leakage — Output',        scope: ['output'],           action: 'rewrite', mode: 'parallel',   managed: true,  enabled: true,  fires24h: 8  },
  { id: 'mgd-003', priority: 3,  name: 'Toxicity / Hate Speech',      scope: ['input', 'output'],  action: 'block',   mode: 'parallel',   managed: true,  enabled: true,  fires24h: 3  },
  { id: 'mgd-004', priority: 4,  name: 'Topic Restriction',           scope: ['input'],            action: 'tag',     mode: 'parallel',   managed: true,  enabled: true,  fires24h: 21 },
  { id: 'mgd-005', priority: 5,  name: 'Tool Call Parameter Anomaly', scope: ['tool_call'],        action: 'block',   mode: 'parallel',   managed: true,  enabled: false, fires24h: 0  },
  { id: 'mgd-006', priority: 6,  name: 'Knowledge Grounding Check',  scope: ['output'],           action: 'tag',     mode: 'sequential', managed: true,  enabled: false, fires24h: 0  },
  { id: 'agt-001', priority: 100,name: 'Agent Support Scope',        scope: ['input'],            action: 'block',   mode: 'parallel',   managed: false, enabled: true,  fires24h: 5, agentId: 'AGT-001' },
  { id: 'cst-001', priority: 10, name: 'No Competitor Mentions',      scope: ['output'],           action: 'rewrite', mode: 'parallel',   managed: false, enabled: true,  fires24h: 2  },
  { id: 'cst-002', priority: 11, name: 'Max Response Length',         scope: ['output'],           action: 'rewrite', mode: 'parallel',   managed: false, enabled: true,  fires24h: 5  },
]

export const piiEntries: PiiEntry[] = [
  { token: 'PII_EMAIL_a3f1',  type: 'EMAIL',       agent: 'AGT-001', maskedAt: '14:31:22', ttl: '23h 12m', hits: 3  },
  { token: 'PII_PHONE_b2c8',  type: 'PHONE',       agent: 'AGT-001', maskedAt: '14:31:22', ttl: '23h 12m', hits: 1  },
  { token: 'PII_NAME_d9e4',   type: 'FULL_NAME',   agent: 'AGT-001', maskedAt: '14:29:10', ttl: '23h 09m', hits: 2  },
  { token: 'PII_SSN_f7a2',    type: 'SSN',         agent: 'AGT-002', maskedAt: '14:28:42', ttl: '23h 08m', hits: 1  },
  { token: 'PII_ACCT_c1b9',   type: 'BANK_ACCT',   agent: 'AGT-002', maskedAt: '14:28:42', ttl: '23h 08m', hits: 1  },
  { token: 'PII_EMAIL_g5h3',  type: 'EMAIL',       agent: 'AGT-003', maskedAt: '14:27:30', ttl: '23h 07m', hits: 1  },
  { token: 'PII_ADDR_k2m7',   type: 'ADDRESS',     agent: 'AGT-001', maskedAt: '14:25:11', ttl: '23h 05m', hits: 0  },
  { token: 'PII_DOB_p4q1',    type: 'DATE_OF_BIRTH',agent: 'AGT-002',maskedAt: '14:22:04', ttl: '23h 02m', hits: 0  },
]

export const endpoints: Endpoint[] = [
  { id: 'ep-001', label: 'OpenAI US-East',   model: 'gpt-4o',     region: 'us-east-1', ttftP50: '78ms',  ttftP99: '124ms', avgTps: '46',  errorRate: '0.2%', quotaUsed: 34, circuitState: 'CLOSED',    score: 0.94 },
  { id: 'ep-002', label: 'Anthropic US',     model: 'claude-3.5', region: 'us-east-1', ttftP50: '142ms', ttftP99: '201ms', avgTps: '38',  errorRate: '0.4%', quotaUsed: 61, circuitState: 'CLOSED',    score: 0.87 },
  { id: 'ep-003', label: 'Google US-Central',model: 'gemini-pro', region: 'us-central', ttftP50: '198ms', ttftP99: '412ms', avgTps: '28',  errorRate: '3.1%', quotaUsed: 12, circuitState: 'HALF-OPEN', score: 0.51 },
  { id: 'ep-004', label: 'OpenAI EU-West',   model: 'gpt-4o',     region: 'eu-west-1', ttftP50: '91ms',  ttftP99: '148ms', avgTps: '43',  errorRate: '0.6%', quotaUsed: 22, circuitState: 'CLOSED',    score: 0.90 },
]

export const apiKeys: APIKey[] = [
  { id: 'key-001', name: 'Production Main', prefix: 'sk', tenantId: 'acme-corp', lastUsed: '2 mins ago', isActive: true, createdAt: '2024-03-10' },
  { id: 'key-002', name: 'Support Agent Key', prefix: 'sk', tenantId: 'acme-corp', agentId: 'AGT-001', lastUsed: '1 hour ago', isActive: true, createdAt: '2024-03-12' },
  { id: 'key-003', name: 'Dev Test Key', prefix: 'sk', tenantId: 'acme-corp', lastUsed: 'Never', isActive: false, createdAt: '2024-03-15' },
]

export const costData = [0.21,0.16,0.11,0.08,0.05,0.08,0.21,0.37,0.55,0.74,0.84,0.79,0.76,0.82,0.71,0.63,0.74,0.92,1.0,0.84,0.76,0.63,0.47,0.32]
