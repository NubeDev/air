import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { 
  Play, 
  Database, 
  FileText, 
  CheckCircle, 
  Circle,
  ArrowRight,
  ArrowLeft
} from 'lucide-react';
import { SessionStart } from '@/components/workflow/SessionStart';
import { SchemaViewer } from '@/components/workflow/SchemaViewer';
import { ScopeBuilder } from '@/components/workflow/ScopeBuilder';

type WorkflowStep = 
  | 'session_start'
  | 'learn'
  | 'scope_build'
  | 'scope_approve'
  | 'query_generate'
  | 'query_approve'
  | 'api_create'
  | 'execute';

interface WorkflowState {
  currentStep: WorkflowStep;
  sessionId?: string;
  sessionType?: 'file' | 'database';
  schemaData?: any;
  scopeData?: any;
  queryData?: any;
  apiData?: any;
}

const workflowSteps = [
  { id: 'session_start', title: 'Start Session', description: 'Upload file or connect to database' },
  { id: 'learn', title: 'Learn Structure', description: 'Discover schema and sample data' },
  { id: 'scope_build', title: 'Build Scope', description: 'Define analysis requirements' },
  { id: 'scope_approve', title: 'Approve Scope', description: 'Review and approve scope' },
  { id: 'query_generate', title: 'Generate Query', description: 'Create SQL or execution plan' },
  { id: 'query_approve', title: 'Approve Query', description: 'Review generated query' },
  { id: 'api_create', title: 'Create API', description: 'Save as executable API' },
  { id: 'execute', title: 'Execute', description: 'Run at scale and get results' },
];

export function WorkflowPage() {
  const [workflowState, setWorkflowState] = useState<WorkflowState>({
    currentStep: 'session_start'
  });

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
            onScopeBuilt={handleStepComplete}
          />
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
          <h1 className="text-3xl font-bold text-gray-900 mb-2">AIR Workflow</h1>
          <p className="text-gray-600">Transform your data into actionable insights through our guided workflow</p>
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
              disabled={currentStepIndex === workflowSteps.length - 1}
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
