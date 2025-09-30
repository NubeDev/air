# AIR UI Specification

## Overview
A modern, ChatGPT-style web interface for the AIR (AI Intelligence Reporting) system built with React, shadcn/ui, and Tailwind CSS. The UI provides an intuitive way to create reports, interact with AI, and query data using schema-driven forms.

## Tech Stack
- **Frontend**: React 18 + TypeScript
- **UI Library**: shadcn/ui + Tailwind CSS
- **State Management**: Zustand
- **HTTP Client**: Axios
- **Form Handling**: React Hook Form + Zod
- **Icons**: Lucide React
- **Build Tool**: Vite
- **Backend**: Go (with CORS enabled)

## Project Structure
```
air-ui/
├── src/
│   ├── components/
│   │   ├── ui/                 # shadcn/ui components
│   │   ├── layout/             # Layout components
│   │   ├── chat/               # Chat interface components
│   │   ├── reports/            # Report management components
│   │   ├── forms/              # Schema-driven form components
│   │   └── common/             # Shared components
│   ├── hooks/                  # Custom React hooks
│   ├── lib/                    # Utilities and configurations
│   ├── stores/                 # Zustand stores
│   ├── types/                  # TypeScript type definitions
│   ├── services/               # API service functions
│   └── pages/                  # Page components
├── public/
├── package.json
├── tailwind.config.js
├── tsconfig.json
└── vite.config.ts
```

## Core Features

### 1. Main Dashboard
- **Layout**: Sidebar navigation + main content area
- **Navigation**: Reports, Chat, Settings, Help
- **Header**: User info, notifications, theme toggle
- **Quick Actions**: New Report, New Chat, Upload File

### 2. Report Management
- **Report List**: Grid/list view of all reports
- **Report Cards**: Show title, description, last run, status
- **Actions**: Edit, Delete, Duplicate, Export
- **Search & Filter**: By name, date, status, tags

### 3. Chat Interface (ChatGPT-style)
- **Chat Window**: Full-height conversation area
- **Message Types**: User, AI, System
- **Input Area**: Text input with send button
- **Message Actions**: Copy, Regenerate, Edit
- **Typing Indicator**: Show when AI is responding
- **Message History**: Scrollable conversation

### 4. Schema-Driven Forms
- **Dynamic Form Generation**: Based on report schema
- **Field Types**: Text, Select, Date, Number, Checkbox
- **Validation**: Real-time validation using Zod
- **Auto-complete**: For enum fields
- **Form State**: Draft saving, validation errors

### 5. Data Visualization
- **Query Results**: Table view with pagination
- **Charts**: Basic charts for numeric data
- **Export Options**: CSV, JSON, PDF
- **Data Actions**: Sort, filter, search

## Component Architecture

### 1. Layout Components
```typescript
// MainLayout.tsx
interface MainLayoutProps {
  children: React.ReactNode;
}

// Sidebar.tsx
interface SidebarProps {
  isCollapsed: boolean;
  onToggle: () => void;
}

// Header.tsx
interface HeaderProps {
  user: User;
  notifications: Notification[];
}
```

### 2. Chat Components
```typescript
// ChatWindow.tsx
interface ChatWindowProps {
  reportId?: string;
  onReportCreated: (report: Report) => void;
}

// Message.tsx
interface MessageProps {
  message: ChatMessage;
  onRegenerate?: () => void;
  onEdit?: (content: string) => void;
}

// ChatInput.tsx
interface ChatInputProps {
  onSend: (message: string) => void;
  disabled?: boolean;
  placeholder?: string;
}
```

### 3. Report Components
```typescript
// ReportList.tsx
interface ReportListProps {
  reports: Report[];
  onSelect: (report: Report) => void;
  onDelete: (id: string) => void;
}

// ReportCard.tsx
interface ReportCardProps {
  report: Report;
  onSelect: () => void;
  onEdit: () => void;
  onDelete: () => void;
}

// ReportForm.tsx
interface ReportFormProps {
  report?: Report;
  onSubmit: (data: ReportFormData) => void;
  onCancel: () => void;
}
```

### 4. Schema Form Components
```typescript
// SchemaForm.tsx
interface SchemaFormProps {
  schema: JSONSchema;
  onSubmit: (data: Record<string, any>) => void;
  initialData?: Record<string, any>;
}

// DynamicField.tsx
interface DynamicFieldProps {
  field: SchemaField;
  value: any;
  onChange: (value: any) => void;
  error?: string;
}

// FieldRenderer.tsx
interface FieldRendererProps {
  field: SchemaField;
  value: any;
  onChange: (value: any) => void;
}
```

## State Management (Zustand)

### 1. Chat Store
```typescript
interface ChatStore {
  messages: ChatMessage[];
  isLoading: boolean;
  currentReportId: string | null;
  addMessage: (message: ChatMessage) => void;
  clearMessages: () => void;
  setLoading: (loading: boolean) => void;
  setCurrentReport: (reportId: string | null) => void;
}
```

### 2. Reports Store
```typescript
interface ReportsStore {
  reports: Report[];
  selectedReport: Report | null;
  isLoading: boolean;
  fetchReports: () => Promise<void>;
  selectReport: (report: Report) => void;
  createReport: (data: CreateReportData) => Promise<Report>;
  updateReport: (id: string, data: UpdateReportData) => Promise<void>;
  deleteReport: (id: string) => Promise<void>;
}
```

### 3. UI Store
```typescript
interface UIStore {
  sidebarCollapsed: boolean;
  theme: 'light' | 'dark';
  notifications: Notification[];
  toggleSidebar: () => void;
  setTheme: (theme: 'light' | 'dark') => void;
  addNotification: (notification: Notification) => void;
  removeNotification: (id: string) => void;
}
```

## API Integration

### 1. API Service
```typescript
// api/reports.ts
export const reportsApi = {
  list: () => api.get<Report[]>('/v1/reports'),
  get: (id: string) => api.get<Report>(`/v1/reports/${id}`),
  create: (data: CreateReportData) => api.post<Report>('/v1/reports', data),
  update: (id: string, data: UpdateReportData) => api.put<Report>(`/v1/reports/${id}`, data),
  delete: (id: string) => api.delete(`/v1/reports/${id}`),
  execute: (id: string, params: Record<string, any>) => api.post<ReportRun>(`/v1/reports/${id}/execute`, { params }),
  getSchema: (id: string) => api.get<ReportSchema>(`/v1/reports/${id}/schema`),
  getData: (id: string) => api.get<ReportData>(`/v1/reports/${id}/data`),
};

// api/chat.ts
export const chatApi = {
  sendMessage: (message: string, reportId?: string) => api.post<ChatResponse>('/v1/chat', { message, report_id: reportId }),
  getHistory: (reportId: string) => api.get<ChatMessage[]>(`/v1/chat/history/${reportId}`),
};
```

### 2. Type Definitions
```typescript
// types/api.ts
interface Report {
  id: string;
  key: string;
  title: string;
  description?: string;
  owner: string;
  created_at: string;
  updated_at: string;
  archived: boolean;
}

interface ReportSchema {
  report_id: string;
  schema: JSONSchema;
}

interface ReportRun {
  id: string;
  report_id: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  row_count: number;
  results: any[];
  started_at: string;
  finished_at: string;
  error_text?: string;
}

interface ChatMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: string;
  report_id?: string;
}
```

## UI/UX Design

### 1. Color Scheme
- **Primary**: Blue (shadcn/ui default)
- **Secondary**: Gray
- **Success**: Green
- **Warning**: Yellow
- **Error**: Red
- **Background**: White/Dark mode

### 2. Typography
- **Font**: Inter (shadcn/ui default)
- **Headings**: Font weights 600-700
- **Body**: Font weight 400
- **Code**: JetBrains Mono

### 3. Spacing
- **Base unit**: 4px (Tailwind default)
- **Component spacing**: 16px, 24px, 32px
- **Page margins**: 24px, 32px, 48px

### 4. Responsive Design
- **Mobile**: 320px - 768px
- **Tablet**: 768px - 1024px
- **Desktop**: 1024px+

## Implementation Phases

### Phase 1: Core Setup (Week 1)
- [ ] Project initialization with Vite + React + TypeScript
- [ ] shadcn/ui setup and configuration
- [ ] Basic layout components (Sidebar, Header, MainLayout)
- [ ] Routing setup with React Router
- [ ] Basic state management with Zustand

### Phase 2: Chat Interface (Week 2)
- [ ] Chat window component
- [ ] Message components (User, AI, System)
- [ ] Chat input with send functionality
- [ ] Message history and persistence
- [ ] Typing indicators and loading states

### Phase 3: Report Management (Week 3)
- [ ] Report list and grid views
- [ ] Report cards with actions
- [ ] Report creation and editing forms
- [ ] Report deletion and confirmation
- [ ] Search and filtering

### Phase 4: Schema Forms (Week 4)
- [ ] Dynamic form generation from JSON Schema
- [ ] Field renderers for different types
- [ ] Form validation with Zod
- [ ] Auto-complete for enum fields
- [ ] Form state management

### Phase 5: Data Visualization (Week 5)
- [ ] Query results table
- [ ] Basic chart components
- [ ] Export functionality
- [ ] Data pagination and sorting
- [ ] Result actions and filters

### Phase 6: Polish & Testing (Week 6)
- [ ] Error handling and loading states
- [ ] Accessibility improvements
- [ ] Performance optimization
- [ ] Testing with Jest and React Testing Library
- [ ] Documentation and deployment

## Backend Requirements

### 1. CORS Configuration
```go
// Add to Go backend
func setupCORS() gin.HandlerFunc {
    return cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    })
}
```

### 2. Additional Endpoints
- `GET /v1/chat/history/{report_id}` - Get chat history for a report
- `POST /v1/chat` - Send chat message
- `GET /v1/reports/{id}/export` - Export report data
- `POST /v1/reports/{id}/duplicate` - Duplicate a report

## File Structure Example
```
air-ui/
├── src/
│   ├── components/
│   │   ├── ui/
│   │   │   ├── button.tsx
│   │   │   ├── input.tsx
│   │   │   ├── card.tsx
│   │   │   ├── dialog.tsx
│   │   │   ├── form.tsx
│   │   │   ├── select.tsx
│   │   │   ├── textarea.tsx
│   │   │   ├── table.tsx
│   │   │   ├── badge.tsx
│   │   │   ├── avatar.tsx
│   │   │   ├── dropdown-menu.tsx
│   │   │   ├── toast.tsx
│   │   │   ├── sheet.tsx
│   │   │   ├── tabs.tsx
│   │   │   ├── scroll-area.tsx
│   │   │   ├── separator.tsx
│   │   │   ├── skeleton.tsx
│   │   │   ├── spinner.tsx
│   │   │   └── index.ts
│   │   ├── layout/
│   │   │   ├── MainLayout.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   ├── Header.tsx
│   │   │   ├── Navigation.tsx
│   │   │   └── index.ts
│   │   ├── chat/
│   │   │   ├── ChatWindow.tsx
│   │   │   ├── Message.tsx
│   │   │   ├── ChatInput.tsx
│   │   │   ├── MessageList.tsx
│   │   │   ├── TypingIndicator.tsx
│   │   │   └── index.ts
│   │   ├── reports/
│   │   │   ├── ReportList.tsx
│   │   │   ├── ReportCard.tsx
│   │   │   ├── ReportForm.tsx
│   │   │   ├── ReportDetails.tsx
│   │   │   ├── ReportActions.tsx
│   │   │   └── index.ts
│   │   ├── forms/
│   │   │   ├── SchemaForm.tsx
│   │   │   ├── DynamicField.tsx
│   │   │   ├── FieldRenderer.tsx
│   │   │   ├── FormValidation.tsx
│   │   │   └── index.ts
│   │   ├── data/
│   │   │   ├── DataTable.tsx
│   │   │   ├── DataChart.tsx
│   │   │   ├── DataExport.tsx
│   │   │   ├── DataFilters.tsx
│   │   │   └── index.ts
│   │   └── common/
│   │       ├── LoadingSpinner.tsx
│   │       ├── ErrorBoundary.tsx
│   │       ├── EmptyState.tsx
│   │       ├── ConfirmDialog.tsx
│   │       └── index.ts
│   ├── hooks/
│   │   ├── useApi.ts
│   │   ├── useChat.ts
│   │   ├── useReports.ts
│   │   ├── useSchema.ts
│   │   ├── useLocalStorage.ts
│   │   └── index.ts
│   ├── lib/
│   │   ├── api.ts
│   │   ├── utils.ts
│   │   ├── constants.ts
│   │   ├── validations.ts
│   │   └── index.ts
│   ├── stores/
│   │   ├── chatStore.ts
│   │   ├── reportsStore.ts
│   │   ├── uiStore.ts
│   │   ├── authStore.ts
│   │   └── index.ts
│   ├── types/
│   │   ├── api.ts
│   │   ├── chat.ts
│   │   ├── reports.ts
│   │   ├── schema.ts
│   │   ├── ui.ts
│   │   └── index.ts
│   ├── services/
│   │   ├── reportsApi.ts
│   │   ├── chatApi.ts
│   │   ├── schemaApi.ts
│   │   ├── dataApi.ts
│   │   └── index.ts
│   ├── pages/
│   │   ├── Dashboard.tsx
│   │   ├── Reports.tsx
│   │   ├── Chat.tsx
│   │   ├── Settings.tsx
│   │   ├── Help.tsx
│   │   └── index.ts
│   ├── App.tsx
│   ├── main.tsx
│   └── vite-env.d.ts
├── public/
│   ├── favicon.ico
│   ├── logo.svg
│   └── manifest.json
├── package.json
├── tailwind.config.js
├── tsconfig.json
├── vite.config.ts
├── .env.example
├── .gitignore
├── README.md
└── CHANGELOG.md
```

## Getting Started

### 1. Initialize Project
```bash
npm create vite@latest air-ui -- --template react-ts
cd air-ui
npm install
```

### 2. Install Dependencies
```bash
npm install @radix-ui/react-* lucide-react
npm install -D tailwindcss postcss autoprefixer
npm install -D @types/node
```

### 3. Setup shadcn/ui
```bash
npx shadcn-ui@latest init
npx shadcn-ui@latest add button input card dialog form select textarea table badge avatar dropdown-menu toast sheet tabs scroll-area separator skeleton
```

### 4. Install Additional Dependencies
```bash
npm install zustand axios react-hook-form @hookform/resolvers zod
npm install react-router-dom
npm install -D @types/react-router-dom
```

This specification provides a comprehensive foundation for building a modern, scalable UI for the AIR system. The ChatGPT-style interface combined with schema-driven forms will provide an intuitive user experience for creating and querying reports.
