import { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { CheckCircle, ArrowRight, ArrowLeft } from 'lucide-react';
import { SessionStart } from '../components/workflow/SessionStart';
import { SchemaViewer } from '../components/workflow/SchemaViewer';
import { ScopeBuilder } from '../components/workflow/ScopeBuilder';
import { chatApi } from '@/services/chatApi';

type WorkflowStep = 
  | 'session_start'
  | 'learn'
  | 'scope_build'
  | 'query_generate'
  | 'api_create'
  | 'execute';

interface WorkflowState {
  currentStep: WorkflowStep;
  sessionId?: string;
  sessionType?: 'file' | 'database';
  schemaData?: any;
  scopeData?: any;
  scopeText?: string;
  scopeId?: number;
  scopeVersionId?: number;
  queryData?: any;
  apiData?: any;
}

// Child component: Generate SQL query via IR + SQLCoder (isolates hook usage)
function QueryGenerateStep({
  scopeVersionId,
  defaultDatasourceId = 'sqlite-dev',
  defaultSqlModelId,
  sqlModels = [],
  onGenerated,
}: {
  scopeVersionId?: number;
  defaultDatasourceId?: string;
  defaultSqlModelId?: string;
  sqlModels?: Array<{ id: string; provider: string; name: string; capabilities: string[] }>;
  onGenerated: (data: { queryData: { sql: string; datasource_id: string } }) => void;
}) {
  const [datasourceId, setDatasourceId] = useState<string>(defaultDatasourceId);
  const [generating, setGenerating] = useState(false);
  const [sql, setSql] = useState<string>('');
  const [error, setError] = useState<string | null>(null);
  const [sqlModel, setSqlModel] = useState<string>(defaultSqlModelId || '');

  const handleGenerate = async () => {
    if (!scopeVersionId) {
      setError('No approved scope available');
      return;
    }
    try {
      setGenerating(true);
      setError(null);

      const irResp = await fetch('/v1/ir/build', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          scope_version_id: scopeVersionId,
          datasource_id: datasourceId,
        }),
      });
      if (!irResp.ok) {
        let msg = `Build IR failed: ${irResp.statusText}`;
        try {
          const j = await irResp.json();
          if (j?.details) msg += ` — ${j.details}`;
          if (j?.error) msg += ` — ${j.error}`;
        } catch {}
        throw new Error(msg);
      }
      const irJson = await irResp.json();
      const ir = irJson.ir;

      const sqlResp = await fetch('/v1/sql', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ir, datasource_id: datasourceId, model: sqlModel }),
      });
      if (!sqlResp.ok) {
        let msg = `Generate SQL failed: ${sqlResp.statusText}`;
        try {
          const j = await sqlResp.json();
          if (j?.details) msg += ` — ${j.details}`;
          if (j?.error) msg += ` — ${j.error}`;
        } catch {}
        throw new Error(msg);
      }
      const sqlJson = await sqlResp.json();
      setSql(sqlJson.sql || '');
      onGenerated({ queryData: { sql: sqlJson.sql, datasource_id: datasourceId } });
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to generate SQL');
    } finally {
      setGenerating(false);
    }
  };

  return (
    <Card className="p-6">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-2xl font-bold">Generate Query</CardTitle>
        <Badge variant="secondary">AI SQL</Badge>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Datasource ID</label>
            <input
              className="border rounded px-3 py-2 text-sm w-full"
              value={datasourceId}
              onChange={(e) => setDatasourceId(e.target.value)}
              placeholder="sqlite-dev"
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">SQL Model</label>
            <select
              className="border rounded px-3 py-2 text-sm w-full bg-white"
              value={sqlModel}
              onChange={(e) => setSqlModel(e.target.value)}
            >
              {sqlModels.map((m) => (
                <option key={m.id} value={m.id}>{m.provider}: {m.name}</option>
              ))}
            </select>
          </div>
        </div>
        {error && <div className="text-destructive text-sm">{error}</div>}
        <div className="flex justify-end">
          <Button onClick={handleGenerate} disabled={generating}>
            {generating ? 'Generating…' : 'Generate SQL'}
          </Button>
        </div>
        {sql && (
          <Card className="p-4 bg-muted/50">
            <pre className="text-xs whitespace-pre-wrap overflow-x-auto">{sql}</pre>
          </Card>
        )}
      </CardContent>
    </Card>
  );
}

// Child component: Create API (Report + Version)
function ApiCreateStep({
  scopeVersionId,
  sql,
  datasourceId,
  onCreated,
}: {
  scopeVersionId?: number;
  sql?: string;
  datasourceId: string;
  onCreated: (data: { reportId: number; versionId: number }) => void;
}) {
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [reportId, setReportId] = useState<number | null>(null);
  const [versionId, setVersionId] = useState<number | null>(null);

  const handleCreate = async () => {
    if (!scopeVersionId || !sql) {
      setError('Missing scope or SQL. Complete previous steps first.');
      return;
    }
    try {
      setCreating(true);
      setError(null);
      // 1) Create report
      const key = `report_${Date.now()}`;
      const title = `AIR Report ${new Date().toLocaleString()}`;
      const repResp = await fetch('/v1/reports', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ key, title }),
      });
      if (!repResp.ok) throw new Error(`Create report failed: ${repResp.statusText}`);
      const report = await repResp.json();
      setReportId(report.id);

      // 2) Create report version with SQL
      const defJSON = JSON.stringify({ sql });
      const verResp = await fetch(`/v1/reports/${report.id}/versions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ scope_version_id: scopeVersionId, datasource_id: datasourceId, def_json: defJSON }),
      });
      if (!verResp.ok) throw new Error(`Create report version failed: ${verResp.statusText}`);
      const version = await verResp.json();
      setVersionId(version.id);

      onCreated({ reportId: report.id, versionId: version.id });
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to create API');
    } finally {
      setCreating(false);
    }
  };

  return (
    <Card className="p-6">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-2xl font-bold">Create API</CardTitle>
        <Badge variant="secondary">Executable API</Badge>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="text-sm text-muted-foreground">Save as executable API</div>
        {error && <div className="text-destructive text-sm">{error}</div>}
        <div className="flex items-center gap-3">
          <Button onClick={handleCreate} disabled={creating}>
            {creating ? 'Creating…' : 'Create API'}
          </Button>
          {reportId && (
            <Badge variant="outline" className="text-xs">Report #{reportId}{versionId ? ` v${versionId}` : ''}</Badge>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

// Child component: Execute API with JSON Schema form
function ExecuteStep({ reportId }: { reportId?: number }) {
  const [schema, setSchema] = useState<any | null>(null);
  const [params, setParams] = useState<Record<string, any>>({});
  const [running, setRunning] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [results, setResults] = useState<any>(null);

  useEffect(() => {
    const load = async () => {
      if (!reportId) return;
      try {
        const res = await fetch(`/v1/reports/${reportId}/schema`);
        if (res.ok) {
          const j = await res.json();
          setSchema(j.schema || null);
        }
      } catch {}
    };
    load();
  }, [reportId]);

  const handleChange = (key: string, value: any) => {
    setParams(prev => ({ ...prev, [key]: value }));
  };

  const handleExecute = async () => {
    if (!reportId) return;
    try {
      setRunning(true);
      setError(null);
      const resp = await fetch(`/v1/reports/${reportId}/execute`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ params }),
      });
      if (!resp.ok) throw new Error(`Execute failed: ${resp.statusText}`);
      const data = await resp.json();
      setResults(data);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to execute');
    } finally {
      setRunning(false);
    }
  };

  const properties = schema?.properties || {};
  const keys = Object.keys(properties);

  return (
    <Card className="p-6">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-2xl font-bold">Execute</CardTitle>
        <Badge variant="secondary">Run & Inspect</Badge>
      </CardHeader>
      <CardContent className="space-y-4">
        {schema ? (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {keys.map((k) => {
              const p = properties[k] as any;
              if (p && p.enum) {
                return (
                  <div key={k} className="space-y-2">
                    <label className="text-sm font-medium">{p.title || k}</label>
                    <select
                      className="border rounded px-3 py-2 text-sm w-full bg-white"
                      value={params[k] ?? ''}
                      onChange={(e) => handleChange(k, e.target.value)}
                    >
                      <option value="">—</option>
                      {p.enum.map((v: string) => (
                        <option key={v} value={v}>{v}</option>
                      ))}
                    </select>
                  </div>
                );
              }
              return (
                <div key={k} className="space-y-2">
                  <label className="text-sm font-medium">{p?.title || k}</label>
                  <input
                    className="border rounded px-3 py-2 text-sm w-full"
                    type={p?.format === 'date' ? 'date' : (p?.type === 'number' || p?.type === 'integer') ? 'number' : 'text'}
                    value={params[k] ?? ''}
                    onChange={(e) => handleChange(k, e.target.value)}
                    placeholder={p?.description || ''}
                  />
                </div>
              );
            })}
          </div>
        ) : (
          <div className="text-sm text-muted-foreground">No schema available; you can still execute with empty parameters.</div>
        )}

        {error && <div className="text-destructive text-sm">{error}</div>}
        <div className="flex justify-end">
          <Button onClick={handleExecute} disabled={running || !reportId}>{running ? 'Executing…' : 'Run'}</Button>
        </div>
        {results && (
          <Card className="p-4 bg-muted/50">
            <pre className="text-xs whitespace-pre-wrap overflow-x-auto">{JSON.stringify(results, null, 2)}</pre>
          </Card>
        )}
      </CardContent>
    </Card>
  );
}

const workflowSteps = [
  { id: 'session_start', title: 'Start Session', description: 'Upload file or connect to database' },
  { id: 'learn', title: 'Learn Structure', description: 'Discover schema and sample data' },
  { id: 'scope_build', title: 'Build Scope', description: 'Define analysis requirements' },
  { id: 'query_generate', title: 'Generate Query', description: 'Create SQL or execution plan' },
  { id: 'api_create', title: 'Create API', description: 'Save as executable API' },
  { id: 'execute', title: 'Execute', description: 'Run at scale and get results' },
];

export function WorkflowPage() {
  const [workflowState, setWorkflowState] = useState<WorkflowState>({
    currentStep: 'session_start'
  });
  const [chatModel, setChatModel] = useState<string>('');
  const [models, setModels] = useState<Array<{ id: string; provider: string; name: string; capabilities: string[] }>>([]);
  const [defaultDatasourceId] = useState<string>('sqlite-dev');
  const [defaultSqlModelId, setDefaultSqlModelId] = useState<string>('');

  useEffect(() => {
    (async () => {
      try {
        const res = await chatApi.getModelStatus();
        const ms = res.data.models || [];
        setModels(ms);
        const defaults = (res.data as any).defaults || {};
        if (!chatModel && defaults.chat) setChatModel(defaults.chat);
        if (!defaultSqlModelId && defaults.sql) setDefaultSqlModelId(defaults.sql);
      } catch (e) {
        // ignore
      }
    })();
  }, []);

  const currentStepIndex = workflowSteps.findIndex(step => step.id === workflowState.currentStep);
  const currentStep = workflowSteps[currentStepIndex];

  const handleStepComplete = (stepData: any) => {
    setWorkflowState(prev => ({
      ...prev,
      ...stepData
    }));
  };

  const handleNextStep = () => {
    if (currentStepIndex < workflowSteps.length - 1) {
      const nextStep = workflowSteps[currentStepIndex + 1];
      setWorkflowState(prev => ({
        ...prev,
        currentStep: nextStep.id as WorkflowStep
      }));
    }
  };

  const handlePrevStep = () => {
    if (currentStepIndex > 0) {
      const prevStep = workflowSteps[currentStepIndex - 1];
      setWorkflowState(prev => ({
        ...prev,
        currentStep: prevStep.id as WorkflowStep
      }));
    }
  };

  const canGoNext = (() => {
    if (currentStepIndex === workflowSteps.length - 1) return false;
    if (workflowState.currentStep === 'scope_build') {
      return !!workflowState.scopeVersionId; // require saved scope version
    }
    return true;
  })();

  const renderCurrentStep = () => {
    switch (workflowState.currentStep) {
      case 'session_start':
        return (
          <SessionStart 
            onSessionStart={handleStepComplete}
            onStepComplete={handleNextStep}
            sessionType={workflowState.sessionType}
          />
        );
      case 'learn':
        return (
          <SchemaViewer 
            sessionId={workflowState.sessionId}
            sessionType={workflowState.sessionType}
            onSchemaLoaded={handleStepComplete}
          />
        );
      case 'scope_build':
        return (
          <ScopeBuilder 
            sessionId={workflowState.sessionId}
            schemaData={workflowState.schemaData}
            chatModel={chatModel as any}
            onScopeBuilt={(d) => {
              handleStepComplete(d);
              handleNextStep();
            }}
          />
        );
      case 'query_generate':
        return (
          <QueryGenerateStep
            scopeVersionId={workflowState.scopeVersionId}
            defaultDatasourceId={defaultDatasourceId}
            defaultSqlModelId={defaultSqlModelId}
            sqlModels={models.filter(m => m.capabilities.includes('sql'))}
            onGenerated={(d: { queryData: { sql: string; datasource_id: string } }) => handleStepComplete(d)}
          />
        );
      case 'api_create':
        return (
          <ApiCreateStep
            scopeVersionId={workflowState.scopeVersionId}
            sql={workflowState.queryData?.sql}
            datasourceId={defaultDatasourceId}
            onCreated={({ reportId }) => handleStepComplete({ apiData: { reportId } })}
          />
        );
      case 'execute':
        return (
          <ExecuteStep reportId={workflowState.apiData?.reportId} />
        );
      default:
        return (
          <Card>
            <CardContent className="p-6">
              <div className="text-center text-muted-foreground">
                <h3 className="text-lg font-semibold mb-2">{currentStep.title}</h3>
                <p>{currentStep.description}</p>
                <p className="text-sm mt-4">Coming soon...</p>
              </div>
            </CardContent>
          </Card>
        );
    }
  };

  return (
    <div className="min-h-full bg-white p-6">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 mb-2">AIR Workflow</h1>
              <p className="text-gray-600">Transform your data into actionable insights through our guided workflow</p>
            </div>
            <div className="flex items-center space-x-3">
              <label className="text-sm text-muted-foreground">Chat Model</label>
              <select
                className="border rounded px-3 py-2 text-sm bg-white"
                value={chatModel}
                onChange={(e) => setChatModel(e.target.value)}
              >
                {models.filter(m => m.capabilities.includes('chat')).map(m => (
                  <option key={m.id} value={m.id}>{m.provider}: {m.name}</option>
                ))}
              </select>
            </div>
          </div>
        </div>

        {/* Progress Steps */}
        <Card className="mb-8">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              {workflowSteps.map((step, index) => {
                const isActive = step.id === workflowState.currentStep;
                const isCompleted = index < currentStepIndex;
                const isAccessible = index <= currentStepIndex;

                return (
                  <div key={step.id} className="flex items-center">
                    <div className="flex flex-col items-center">
                      <div className={`
                        w-10 h-10 rounded-full flex items-center justify-center text-sm font-medium
                        ${isCompleted 
                          ? 'bg-primary text-primary-foreground' 
                          : isActive 
                            ? 'bg-primary text-primary-foreground ring-4 ring-primary/20' 
                            : isAccessible
                              ? 'bg-muted text-muted-foreground'
                              : 'bg-gray-100 text-gray-400'
                        }
                      `}>
                        {isCompleted ? (
                          <CheckCircle className="h-5 w-5" />
                        ) : (
                          <span>{index + 1}</span>
                        )}
                      </div>
                      <div className="mt-2 text-center">
                        <div className={`text-sm font-medium ${isActive ? 'text-primary' : 'text-muted-foreground'}`}>
                          {step.title}
                        </div>
                        <div className="text-xs text-muted-foreground mt-1 max-w-24">
                          {step.description}
                        </div>
                      </div>
                    </div>
                    {index < workflowSteps.length - 1 && (
                      <div className={`flex-1 h-0.5 mx-4 ${isCompleted ? 'bg-primary' : 'bg-muted'}`} />
                    )}
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>

        {/* Current Step Content */}
        <div className="mb-6">
          {renderCurrentStep()}
        </div>

        {/* Navigation */}
        <div className="flex justify-between">
          <Button
            variant="outline"
            onClick={handlePrevStep}
            disabled={currentStepIndex === 0}
            className="flex items-center"
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            Previous
          </Button>
          
          <div className="flex items-center space-x-4">
            <Badge variant="outline" className="text-sm">
              Step {currentStepIndex + 1} of {workflowSteps.length}
            </Badge>
            
            <Button
              onClick={handleNextStep}
              disabled={!canGoNext}
              className="flex items-center"
            >
              Next
              <ArrowRight className="h-4 w-4 ml-2" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
