import { Shield, Terminal, Copy, Check } from 'lucide-react'
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

interface IntegrationGuideProps {
  tenantId: string | null
  onClose: () => void
}

export default function IntegrationGuide({ tenantId, onClose }: IntegrationGuideProps) {
  const [copied, setCopied] = useState<string | null>(null)

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text)
    setCopied(id)
    setTimeout(() => setCopied(null), 2000)
  }

  const commands = [
    {
      id: 'env',
      label: 'Set Environment Variable',
      cmd: 'export OPENAI_BASE_URL="http://localhost:8080/v1"',
    },
    {
      id: 'test',
      label: 'Test Connection',
      cmd: 'curl http://localhost:8080/v1/models -H "Authorization: Bearer sk_your_key"',
    },
    {
      id: 'install-py',
      label: 'Install Python SDK',
      cmd: 'pip install openai',
    },
    {
      id: 'install-js',
      label: 'Install Node.js SDK',
      cmd: 'npm install openai',
    },
  ]

  return (
    <DialogContent className="sm:max-w-[1000px] w-[95vw] h-[90vh] overflow-hidden flex flex-col">
      <DialogHeader>
        <DialogTitle className="font-black tracking-tight flex items-center gap-2">
          <Shield className="w-5 h-5 text-primary" /> Integration Guide
        </DialogTitle>
        <DialogDescription className="font-mono text-xs">
          Connect your application to the AI Gateway using OpenAI-compatible SDKs.
        </DialogDescription>
      </DialogHeader>

      <ScrollArea className="flex-1 min-h-0 pr-4">
        <div className="space-y-6 py-4">
          <section>
            <h3 className="font-mono text-[10px] tracking-widest uppercase text-primary mb-3">Quick Start Commands</h3>
            <div className="grid grid-cols-1 gap-3">
              {commands.map((c) => (
                <div key={c.id} className="relative group">
                  <div className="absolute -top-2 left-3 px-1.5 bg-background text-[9px] font-mono text-muted-foreground uppercase tracking-wider z-10">
                    {c.label}
                  </div>
                  <div className="flex items-center gap-2 bg-muted/50 p-3 rounded-lg border font-mono text-xs overflow-hidden">
                    <Terminal className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
                    <code className="flex-1 truncate text-foreground/80">{c.cmd}</code>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-7 w-7 opacity-0 group-hover:opacity-100 transition-opacity"
                      onClick={() => copyToClipboard(c.cmd, c.id)}
                    >
                      {copied === c.id ? <Check className="w-3 h-3 text-green-500" /> : <Copy className="w-3 h-3" />}
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </section>

          <section>
            <h3 className="font-mono text-[10px] tracking-widest uppercase text-primary mb-2">Endpoint Configuration</h3>
            <div className="bg-muted p-3 rounded border font-mono text-xs space-y-2">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Proxy URL</span>
                <span className="text-foreground">http://localhost:8080/v1</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Auth Method</span>
                <span className="text-foreground">Bearer Token</span>
              </div>
            </div>
          </section>

          <section>
            <h3 className="font-mono text-[10px] tracking-widest uppercase text-primary mb-2">Context Headers</h3>
            <p className="text-[11px] text-muted-foreground mb-3 leading-relaxed">
              Pass additional metadata for tracing, multi-tenancy, and routing via standard HTTP headers.
            </p>
            <Table>
              <TableHeader>
                <TableRow className="border-border hover:bg-transparent">
                  <TableHead className="font-mono text-[9px] tracking-widest h-8">HEADER</TableHead>
                  <TableHead className="font-mono text-[9px] tracking-widest h-8">DESCRIPTION</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {[
                  { h: 'X-Agent-Id', d: 'Identifier for tracing and policy' },
                  { h: 'X-Tenant-Id', d: 'Tenant namespace isolation' },
                  { h: 'X-Thread-Id', d: 'Persistent conversation ID' },
                ].map((row) => (
                  <TableRow key={row.h} className="border-border py-0">
                    <TableCell className="font-mono text-[10px] py-2">{row.h}</TableCell>
                    <TableCell className="text-[10px] py-2 text-muted-foreground">{row.d}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </section>

          <section>
            <h3 className="font-mono text-[10px] tracking-widest uppercase text-primary mb-2">Code Example</h3>
            <Tabs defaultValue="python" className="w-full">
              <TabsList className="bg-muted w-full justify-start rounded-b-none border-b-0 h-9 p-0.5">
                <TabsTrigger value="python" className="font-mono text-[10px] h-8 data-[state=active]:bg-background">Python</TabsTrigger>
                <TabsTrigger value="node" className="font-mono text-[10px] h-8 data-[state=active]:bg-background">Node.js</TabsTrigger>
                <TabsTrigger value="curl" className="font-mono text-[10px] h-8 data-[state=active]:bg-background">cURL</TabsTrigger>
                <TabsTrigger value="claude" className="font-mono text-[10px] h-8 data-[state=active]:bg-background">Claude Code</TabsTrigger>
              </TabsList>
              <TabsContent value="python" className="mt-0">
                <div className="relative group">
                  <Button
                    variant="ghost"
                    size="icon"
                    className="absolute top-2 right-2 h-7 w-7 opacity-0 group-hover:opacity-100 transition-opacity bg-background/50 hover:bg-background"
                    onClick={() => copyToClipboard(`from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="sk_your_key_here"
)

response = client.chat.completions.create(
    model="gpt-4",
    messages=[{"role": "user", "content": "Hello!"}],
    extra_headers={
        "X-Agent-Id": "my-agent",
        "X-Tenant-Id": "${tenantId || 'tenant-id'}"
    }
)`, 'code-py')}
                  >
                    {copied === 'code-py' ? <Check className="w-3 h-3 text-green-500" /> : <Copy className="w-3 h-3" />}
                  </Button>
                  <pre className="bg-muted/50 p-4 rounded-b border border-t-0 font-mono text-[11px] leading-relaxed overflow-x-auto">
{`from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="sk_your_key_here"
)

response = client.chat.completions.create(
    model="gpt-4",
    messages=[{"role": "user", "content": "Hello!"}],
    extra_headers={
        "X-Agent-Id": "my-agent",
        "X-Tenant-Id": "${tenantId || 'tenant-id'}"
    }
)`}
                  </pre>
                </div>
              </TabsContent>
              <TabsContent value="node" className="mt-0">
                <div className="relative group">
                  <Button
                    variant="ghost"
                    size="icon"
                    className="absolute top-2 right-2 h-7 w-7 opacity-0 group-hover:opacity-100 transition-opacity bg-background/50 hover:bg-background"
                    onClick={() => copyToClipboard(`import OpenAI from 'openai';

const openai = new OpenAI({
  baseURL: 'http://localhost:8080/v1',
  apiKey: 'sk_your_key_here',
});

const response = await openai.chat.completions.create({
  model: 'gpt-4',
  messages: [{ role: 'user', content: 'Hello!' }],
}, {
  headers: {
    'X-Agent-Id': 'my-agent',
    'X-Tenant-Id': '${tenantId || 'tenant-id'}',
  }
});`, 'code-node')}
                  >
                    {copied === 'code-node' ? <Check className="w-3 h-3 text-green-500" /> : <Copy className="w-3 h-3" />}
                  </Button>
                  <pre className="bg-muted/50 p-4 rounded-b border border-t-0 font-mono text-[11px] leading-relaxed overflow-x-auto">
{`import OpenAI from 'openai';

const openai = new OpenAI({
  baseURL: 'http://localhost:8080/v1',
  apiKey: 'sk_your_key_here',
});

const response = await openai.chat.completions.create({
  model: 'gpt-4',
  messages: [{ role: 'user', content: 'Hello!' }],
}, {
  headers: {
    'X-Agent-Id': 'my-agent',
    'X-Tenant-Id': '${tenantId || 'tenant-id'}',
  }
});`}
                  </pre>
                </div>
              </TabsContent>
              <TabsContent value="curl" className="mt-0">
                <div className="relative group">
                  <Button
                    variant="ghost"
                    size="icon"
                    className="absolute top-2 right-2 h-7 w-7 opacity-0 group-hover:opacity-100 transition-opacity bg-background/50 hover:bg-background"
                    onClick={() => copyToClipboard(`curl http://localhost:8080/v1/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer sk_your_key_here" \\
  -H "X-Agent-Id: my-agent" \\
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'`, 'code-curl')}
                  >
                    {copied === 'code-curl' ? <Check className="w-3 h-3 text-green-500" /> : <Copy className="w-3 h-3" />}
                  </Button>
                  <pre className="bg-muted/50 p-4 rounded-b border border-t-0 font-mono text-[11px] leading-relaxed overflow-x-auto">
{`curl http://localhost:8080/v1/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer sk_your_key_here" \\
  -H "X-Agent-Id: my-agent" \\
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'`}
                  </pre>
                </div>
              </TabsContent>
              <TabsContent value="claude" className="mt-0">
                <div className="relative group">
                  <Button
                    variant="ghost"
                    size="icon"
                    className="absolute top-2 right-2 h-7 w-7 opacity-0 group-hover:opacity-100 transition-opacity bg-background/50 hover:bg-background"
                    onClick={() => copyToClipboard(`export ANTHROPIC_BASE_URL="http://localhost:8080/v1"
export ANTHROPIC_API_KEY="sk_your_gateway_key"
claude`, 'code-claude')}
                  >
                    {copied === 'code-claude' ? <Check className="w-3 h-3 text-green-500" /> : <Copy className="w-3 h-3" />}
                  </Button>
                  <pre className="bg-muted/50 p-4 rounded-b border border-t-0 font-mono text-[11px] leading-relaxed overflow-x-auto">
{`# 1. Route Claude Code CLI through the gateway
export ANTHROPIC_BASE_URL="http://localhost:8080/v1"

# 2. Use your Gateway API Key as the Anthropic Key
export ANTHROPIC_API_KEY="sk_your_gateway_key"

# 3. Launch the CLI
claude`}
                  </pre>
                </div>
              </TabsContent>
            </Tabs>
          </section>

          <section>
            <h3 className="font-mono text-[10px] tracking-widest uppercase text-primary mb-2">Protocol Translation</h3>
            <p className="text-[11px] text-muted-foreground mb-3 leading-relaxed">
              The gateway automatically translates between protocols. It exposes both <strong>OpenAI-compatible</strong> and <strong>Anthropic-native</strong> endpoints.
            </p>
            <div className="bg-primary/5 p-3 rounded-lg border border-primary/10 space-y-3">
              <p className="text-[10px] text-foreground leading-tight">
                <strong>Standard Translation:</strong> Use the <code>openai</code> SDK to talk to any model (Claude, Gemini, etc.) via <code>/v1/chat/completions</code>.
              </p>
              <p className="text-[10px] text-foreground leading-tight border-t border-primary/10 pt-2">
                <strong>Native Passthrough:</strong> Use <code>claude-code</code> or the Anthropic SDK by pointing them to <code>/v1/messages</code>. The gateway transparently maps the <code>x-api-key</code> header for auth and detects the <code>anthropic-version</code> header to return the correct model list format.
              </p>
            </div>
          </section>

          <section>
            <h3 className="font-mono text-[10px] tracking-widest uppercase text-primary mb-2">CLI & IDE Integration</h3>
            <p className="text-[11px] text-muted-foreground mb-4 leading-relaxed font-mono uppercase">
              DROP-IN COMPATIBILITY WITH ANY TOOL SUPPORTING CUSTOM OPENAI ENDPOINTS.
            </p>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div className="p-3 border rounded-lg bg-primary/5">
                <p className="font-mono text-[10px] font-bold text-primary uppercase mb-1">IDE Extensions</p>
                <p className="text-[10px] text-muted-foreground leading-tight">
                  Works with <strong>Cursor</strong>, <strong>Continue</strong>, and <strong>Cline</strong>. Point the OpenAI endpoint to your gateway URL.
                </p>
              </div>
              <div className="p-3 border rounded-lg bg-primary/5">
                <p className="font-mono text-[10px] font-bold text-primary uppercase mb-1">CLI Tools</p>
                <p className="text-[10px] text-muted-foreground leading-tight">
                  Compatible with <strong>aichat</strong>, <strong>mods</strong>, and <strong>claude-code</strong>. Set <code className="bg-muted px-1">OPENAI_BASE_URL</code> or <code className="bg-muted px-1">ANTHROPIC_BASE_URL</code>.
                </p>
              </div>
            </div>
            <div className="mt-4 p-3 border border-dashed rounded-lg bg-muted/20">
              <p className="text-[10px] text-muted-foreground leading-relaxed italic">
                Note: For model-specific CLIs like <strong>claude-code</strong> or <strong>gemini-cli</strong>, use them via an OpenAI-compatible adapter like <strong>litellm proxy</strong> or check if they support custom base URLs for their underlying SDKs.
              </p>
            </div>
          </section>
        </div>
      </ScrollArea>

      <DialogFooter className="mt-4 pt-4 border-t">
        <Button onClick={onClose} className="font-mono text-xs w-full">Close Guide</Button>
      </DialogFooter>
    </DialogContent>
  )
}
