import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Textarea } from '@/components/ui/textarea';
import { 
  Target, 
  CheckCircle,
  Loader2,
  AlertCircle,
  Wand2,
  Save,
  Eye
} from 'lucide-react';

interface ScopeBuilderProps {
  sessionId?: string;
  schemaData?: any;
  chatModel?: 'openai' | 'llama' | 'sqlcoder';
  onScopeBuilt: (data: { scopeData: any; scopeText?: string; scopeId?: number; scopeVersionId?: number }) => void;
}

export function ScopeBuilder({ sessionId, schemaData, chatModel = 'llama', onScopeBuilt }: ScopeBuilderProps) {
  const [scopeText, setScopeText] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [scopeVersions, setScopeVersions] = useState<any[]>([]);
  const [currentVersion, setCurrentVersion] = useState<any>(null);
  const [scopeId, setScopeId] = useState<number | null>(null);

  useEffect(() => {
    if (sessionId) {
      loadScopeVersions();
    }
  }, [sessionId]);

  const loadScopeVersions = async () => {
    // No session-scoped versions API; we update local state after saving versions
    return;
  };

  const generateScope = async () => {
    if (!schemaData) return;

    setIsGenerating(true);
    setError(null);

    try {
      const schemaSummary = summarizeSchema(schemaData);
      const messages = [
        {
          role: 'system',
          content:
            'You are AIR, a data analysis assistant. Given schema context and a short user goal, draft a concise analysis scope in markdown with clear objectives, metrics, dimensions, filters, and assumptions. Do not fabricate columns. Keep it practical.'
        },
        {
          role: 'user',
          content: `Schema Context\n${schemaSummary}\n\nUser Goal: ${scopeText || 'Describe the dataset and propose a starter analysis scope.'}`
        }
      ];

      const resp = await fetch('/v1/ai/chat/completion', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ messages, model: chatModel })
      });
      if (!resp.ok) throw new Error(`Scope generation failed: ${resp.statusText}`);
      const data = await resp.json();
      const content = data?.Message?.Content || data?.message?.content || '';
      if (!content) throw new Error('Model response missing content');
      setScopeText(content);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to generate scope');
    } finally {
      setIsGenerating(false);
    }
  };

  const saveScope = async (): Promise<{ scopeId: number; version: any } | null> => {
    if (!scopeText.trim()) return null;

    setIsSaving(true);
    setError(null);

    try {
      let id = scopeId;
      if (!id) {
        const create = await fetch('/v1/scopes', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ name: `Session ${sessionId || 'N/A'} Scope` })
        });
        if (!create.ok) throw new Error(`Create scope failed: ${create.statusText}`);
        const created = await create.json();
        id = created.id;
        setScopeId(id);
      }

      const ver = await fetch(`/v1/scopes/${id}/version`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ scope_md: scopeText })
      });
      if (!ver.ok) throw new Error(`Create scope version failed: ${ver.statusText}`);
      const version = await ver.json();
      setCurrentVersion(version);
      setScopeVersions((prev) => [version, ...prev]);
      return { scopeId: id!, version };
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save scope');
      return null;
    } finally {
      setIsSaving(false);
    }
  };

  const refineScope = async () => {
    if (!scopeText.trim()) return;

    setIsGenerating(true);
    setError(null);

    try {
      const schemaSummary = summarizeSchema(schemaData);
      const messages = [
        {
          role: 'system',
          content:
            'You are AIR. Refine the provided analysis scope. Keep structure, remove fluff, clarify objectives, metrics, dimensions, filters, assumptions. Do not invent columns. Return markdown only.'
        },
        {
          role: 'user',
          content: `Schema Context\n${schemaSummary}\n\nCurrent Scope:\n${scopeText}`
        }
      ];

      const resp = await fetch('/v1/ai/chat/completion', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ messages, model: chatModel })
      });
      if (!resp.ok) throw new Error(`Scope refinement failed: ${resp.statusText}`);
      const data = await resp.json();
      const content = data?.Message?.Content || data?.message?.content || '';
      if (!content) throw new Error('Model response missing content');
      setScopeText(content);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to refine scope');
    } finally {
      setIsGenerating(false);
    }
  };

  const handleScopeComplete = async () => {
    let versionObj = currentVersion;
    let id = scopeId;
    if (!versionObj) {
      const saved = await saveScope();
      if (!saved) return; // error state already set
      versionObj = saved.version;
      id = saved.scopeId;
    }
    onScopeBuilt({
      scopeData: { scope_md: scopeText, version: versionObj },
      scopeText,
      scopeId: id || undefined,
      scopeVersionId: versionObj?.id,
    });
  };

  function summarizeSchema(schema: any): string {
    if (!schema) return 'No schema provided.';
    if (schema.columns) {
      const cols = schema.columns.map((c: any) => c.name).join(', ');
      const rows = schema.row_count ?? 'Unknown';
      const file = schema.filename ?? 'Unknown file';
      return `File: ${file}\nColumns: ${cols}\nRows: ${rows}`;
    }
    if (schema.tables) {
      const tables = (schema.tables || []).map((t: any) => t.name).join(', ');
      const db = schema.database_type ?? 'database';
      return `Database: ${db}\nTables: ${tables}`;
    }
    return JSON.stringify(schema).slice(0, 800);
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center">
            <Target className="h-5 w-5 mr-2" />
            Scope Builder
            <Badge variant="outline" className="ml-2">
              Interactive
            </Badge>
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Schema Context */}
          {schemaData && (
            <Card className="bg-muted/50">
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Schema Context</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-sm text-muted-foreground">
                  {schemaData.columns ? (
                    <div>
                      <p><strong>File:</strong> {schemaData.filename || 'Unknown'}</p>
                      <p><strong>Columns:</strong> {schemaData.columns.map((col: any) => col.name).join(', ')}</p>
                      <p><strong>Rows:</strong> {schemaData.row_count || 'Unknown'}</p>
                    </div>
                  ) : (
                    <div>
                      <p><strong>Database:</strong> {schemaData.database_type || 'Unknown'}</p>
                      <p><strong>Tables:</strong> {schemaData.tables?.map((table: any) => table.name).join(', ') || 'Unknown'}</p>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          )}

          {/* Scope Editor */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold">Analysis Scope</h3>
              <div className="flex space-x-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={generateScope}
                  disabled={isGenerating || !schemaData}
                  className="flex items-center"
                >
                  {isGenerating ? (
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  ) : (
                    <Wand2 className="h-4 w-4 mr-2" />
                  )}
                  {isGenerating ? 'Generating...' : 'Generate'}
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={refineScope}
                  disabled={isGenerating || !scopeText.trim()}
                  className="flex items-center"
                >
                  {isGenerating ? (
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  ) : (
                    <Wand2 className="h-4 w-4 mr-2" />
                  )}
                  Refine
                </Button>
              </div>
            </div>

            <Textarea
              value={scopeText}
              onChange={(e) => setScopeText(e.target.value)}
              placeholder="Define your analysis scope here. Include:
- What questions you want to answer
- What insights you're looking for
- What data points are most important
- Any specific analysis requirements

Example:
I want to analyze sales performance trends over the last quarter, focusing on:
- Revenue by product category
- Customer acquisition patterns
- Seasonal variations
- Top-performing regions"
              className="min-h-[300px] font-mono text-sm"
            />

            <div className="flex justify-between items-center">
              <div className="text-sm text-muted-foreground">
                {scopeText.length} characters
              </div>
              <div className="flex space-x-2">
                <Button
                  variant="outline"
                  onClick={saveScope}
                  disabled={isSaving || !scopeText.trim()}
                  className="flex items-center"
                >
                  {isSaving ? (
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  ) : (
                    <Save className="h-4 w-4 mr-2" />
                  )}
                  {isSaving ? 'Saving...' : 'Save Draft'}
                </Button>
                <Button
                  onClick={handleScopeComplete}
                  disabled={!scopeText.trim()}
                  className="flex items-center"
                >
                  <CheckCircle className="h-4 w-4 mr-2" />
                  Complete Scope
                </Button>
              </div>
            </div>
          </div>

          {/* Error Display */}
          {error && (
            <div className="flex items-center space-x-2 text-destructive bg-destructive/10 p-3 rounded-md">
              <AlertCircle className="h-4 w-4" />
              <span className="text-sm">{error}</span>
            </div>
          )}

          {/* Scope Versions */}
          {scopeVersions.length > 0 && (
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Scope Versions</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  {scopeVersions.slice(0, 3).map((version, index) => (
                    <div
                      key={version.id || index}
                      className={`flex items-center justify-between p-2 rounded border cursor-pointer transition-colors ${
                        currentVersion?.id === version.id ? 'bg-primary/10 border-primary' : 'hover:bg-muted'
                      }`}
                      onClick={() => {
                        setCurrentVersion(version);
                        setScopeText(version.scope_md || '');
                      }}
                    >
                      <div className="flex items-center space-x-2">
                        <Eye className="h-4 w-4 text-muted-foreground" />
                        <span className="text-sm">
                          Version {version.version_number || index + 1}
                        </span>
                        <Badge variant="outline" className="text-xs">
                          {version.created_at ? new Date(version.created_at).toLocaleDateString() : 'Unknown'}
                        </Badge>
                      </div>
                      <div className="text-xs text-muted-foreground">
                        {version.scope_md?.length || 0} chars
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
