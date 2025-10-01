import { useEffect, useState } from 'react';
import { chatApi } from '@/services/chatApi';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { wsService } from '@/services/websocket';

type ProviderHealth = { connected: boolean; error?: string };

export function SettingsPage() {
  const [models, setModels] = useState<Array<{ id: string; provider: string; name: string; capabilities: string[] }>>([]);
  const [defaults, setDefaults] = useState<Record<string, string>>({});
  const [health, setHealth] = useState<Record<string, ProviderHealth>>({});
  const [pingResponse, setPingResponse] = useState<string>('');
  const [pingModel, setPingModel] = useState<string>('');
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    load();
  }, []);

  async function load() {
    try {
      const status = await chatApi.getModelStatus();
      setDefaults(status.data.defaults || {});
      setModels(status.data.models || []);
      setHealth(status.data.health || {});
      if (!pingModel && status.data.defaults?.chat) setPingModel(status.data.defaults.chat);
    } catch (e) {
      // ignore
    }
  }

  async function setPrimary(capability: 'chat' | 'sql', model: string) {
    setBusy(true);
    try {
      await chatApi.setPrimaryModel(capability, model);
      await load();
    } finally {
      setBusy(false);
    }
  }

  function pingRaw() {
    setPingResponse('');
    wsService.sendMessage({
      type: 'raw_ai_message',
      payload: { content: 'what model are you?', model: pingModel },
    });
    const handler = (msg: any) => {
      setPingResponse(msg.payload?.content || '');
      wsService.offMessage('raw_ai_response', handler as any);
    };
    wsService.onMessage('raw_ai_response', handler as any);
  }

  const chatModels = models.filter(m => m.capabilities.includes('chat'));
  const sqlModels = models.filter(m => m.capabilities.includes('sql'));

  return (
    <div className="p-6 space-y-6">
      <h1 className="text-2xl font-semibold">Settings</h1>

      <Card>
        <CardContent className="p-4 space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm text-muted-foreground">Default Chat Model</div>
              <div className="text-base font-medium break-all">{defaults.chat || '—'}</div>
            </div>
            <Select value={defaults.chat} onValueChange={(v) => setPrimary('chat', v)}>
              <SelectTrigger className="w-72"><SelectValue placeholder="Select model" /></SelectTrigger>
              <SelectContent>
                {chatModels.map((m) => (
                  <SelectItem key={m.id} value={m.id}>{m.provider}: {m.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm text-muted-foreground">Default SQL Model</div>
              <div className="text-base font-medium break-all">{defaults.sql || '—'}</div>
            </div>
            <Select value={defaults.sql} onValueChange={(v) => setPrimary('sql', v)}>
              <SelectTrigger className="w-72"><SelectValue placeholder="Select model" /></SelectTrigger>
              <SelectContent>
                {sqlModels.map((m) => (
                  <SelectItem key={m.id} value={m.id}>{m.provider}: {m.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="p-4 space-y-3">
          <div className="text-base font-medium">Providers Health</div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {['openai','ollama'].map((p) => (
              <div key={p} className="flex items-center justify-between rounded border p-3">
                <div className="font-medium capitalize">{p}</div>
                <div className={`w-2 h-2 rounded-full ${health[p]?.connected ? 'bg-emerald-500' : 'bg-rose-500'}`} title={health[p]?.error || ''} />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="p-4 space-y-3">
          <div className="text-base font-medium">Raw Ping</div>
          <div className="flex items-center gap-3">
            <Select value={pingModel} onValueChange={setPingModel}>
              <SelectTrigger className="w-72"><SelectValue placeholder="Select model" /></SelectTrigger>
              <SelectContent>
                {models.map((m) => (
                  <SelectItem key={m.id} value={m.id}>{m.provider}: {m.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button disabled={!pingModel || busy} onClick={pingRaw}>Ping model</Button>
          </div>
          {pingResponse && (
            <pre className="bg-muted p-3 rounded text-sm whitespace-pre-wrap">{pingResponse}</pre>
          )}
        </CardContent>
      </Card>
    </div>
  );
}


