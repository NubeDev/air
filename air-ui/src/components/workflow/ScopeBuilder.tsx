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
  onScopeBuilt: (data: { scopeData: any }) => void;
}

export function ScopeBuilder({ sessionId, schemaData, onScopeBuilt }: ScopeBuilderProps) {
  const [scopeText, setScopeText] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [scopeVersions, setScopeVersions] = useState<any[]>([]);
  const [currentVersion, setCurrentVersion] = useState<any>(null);

  useEffect(() => {
    if (sessionId) {
      loadScopeVersions();
    }
  }, [sessionId]);

  const loadScopeVersions = async () => {
    if (!sessionId) return;

    try {
      const response = await fetch(`/v1/sessions/${sessionId}/scope/versions`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        setScopeVersions(data.versions || []);
        if (data.versions && data.versions.length > 0) {
          setCurrentVersion(data.versions[0]);
          setScopeText(data.versions[0].scope_md || '');
        }
      }
    } catch (err) {
      console.error('Failed to load scope versions:', err);
    }
  };

  const generateScope = async () => {
    if (!sessionId || !schemaData) return;

    setIsGenerating(true);
    setError(null);

    try {
      const response = await fetch(`/v1/sessions/${sessionId}/scope/generate`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          schema_data: schemaData,
          context: 'Generate a comprehensive analysis scope based on the provided schema'
        }),
      });

      if (!response.ok) {
        throw new Error(`Scope generation failed: ${response.statusText}`);
      }

      const data = await response.json();
      setScopeText(data.scope_md || '');
      
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to generate scope');
    } finally {
      setIsGenerating(false);
    }
  };

  const saveScope = async () => {
    if (!sessionId || !scopeText.trim()) return;

    setIsSaving(true);
    setError(null);

    try {
      const response = await fetch(`/v1/sessions/${sessionId}/scope/save`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          scope_md: scopeText,
          version_notes: 'User created scope'
        }),
      });

      if (!response.ok) {
        throw new Error(`Scope save failed: ${response.statusText}`);
      }

      const data = await response.json();
      setCurrentVersion(data);
      loadScopeVersions(); // Refresh versions
      
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save scope');
    } finally {
      setIsSaving(false);
    }
  };

  const refineScope = async () => {
    if (!sessionId || !scopeText.trim()) return;

    setIsGenerating(true);
    setError(null);

    try {
      const response = await fetch(`/v1/sessions/${sessionId}/scope/refine`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          current_scope: scopeText,
          refinement_instructions: 'Please refine and improve this scope'
        }),
      });

      if (!response.ok) {
        throw new Error(`Scope refinement failed: ${response.statusText}`);
      }

      const data = await response.json();
      setScopeText(data.refined_scope || scopeText);
      
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to refine scope');
    } finally {
      setIsGenerating(false);
    }
  };

  const handleScopeComplete = () => {
    onScopeBuilt({ scopeData: { scope_md: scopeText, version: currentVersion } });
  };

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
