import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Copy, RotateCcw, Edit, Database, Code, BarChart3 } from 'lucide-react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeHighlight from 'rehype-highlight';
import 'highlight.js/styles/github.css';
import type { ChatMessage as ChatMessageType } from '@/types/api';

interface MessageProps {
  message: ChatMessageType;
  onRegenerate?: () => void;
  onEdit?: (content: string) => void;
}

export function Message({ message, onRegenerate, onEdit }: MessageProps) {
  const isUser = message.role === 'user';

  const handleCopy = () => {
    navigator.clipboard.writeText(message.content);
  };

  const renderDataResult = (data: any) => {
    if (!data) return null;

    if (Array.isArray(data)) {
      return (
        <div className="mt-3">
          <div className="flex items-center gap-2 mb-2">
            <Database className="h-4 w-4" />
            <span className="text-sm font-medium">Query Results ({data.length} rows)</span>
          </div>
          <div className="max-h-60 overflow-auto border rounded-md">
            <table className="w-full text-xs">
              <thead className="bg-muted">
                <tr>
                  {data.length > 0 && Object.keys(data[0]).map((key) => (
                    <th key={key} className="px-2 py-1 text-left font-medium">{key}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {data.slice(0, 10).map((row, index) => (
                  <tr key={index} className="border-t">
                    {Object.values(row).map((value, cellIndex) => (
                      <td key={cellIndex} className="px-2 py-1">{String(value)}</td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
            {data.length > 10 && (
              <div className="px-2 py-1 text-xs text-muted-foreground bg-muted">
                ... and {data.length - 10} more rows
              </div>
            )}
          </div>
        </div>
      );
    }

    return (
      <div className="mt-3">
        <div className="flex items-center gap-2 mb-2">
          <BarChart3 className="h-4 w-4" />
          <span className="text-sm font-medium">Data Result</span>
        </div>
        <pre className="text-xs bg-muted p-2 rounded overflow-auto max-h-40">
          {JSON.stringify(data, null, 2)}
        </pre>
      </div>
    );
  };

  const renderSQL = (sql: string) => {
    if (!sql) return null;

    return (
      <div className="mt-3">
        <div className="flex items-center gap-2 mb-2">
          <Code className="h-4 w-4" />
          <span className="text-sm font-medium">Generated SQL</span>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => navigator.clipboard.writeText(sql)}
            className="h-6 px-2"
          >
            <Copy className="h-3 w-3" />
          </Button>
        </div>
        <pre className="text-xs bg-muted p-2 rounded overflow-auto">
          {sql}
        </pre>
      </div>
    );
  };

  return (
    <div className={cn(
      "flex gap-4 py-6",
      isUser ? "justify-end" : "justify-start"
    )}>
      {!isUser && (
        <div className="flex-shrink-0 w-8 h-8 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white text-sm font-semibold shadow-lg">
          AI
        </div>
      )}
      
      <div className={cn(
        "max-w-[80%] rounded-2xl px-4 py-3 shadow-sm",
        isUser 
          ? "bg-blue-500 text-white ml-12" 
          : "bg-gray-100 text-gray-900 mr-12"
      )}>
        <div className="prose prose-sm max-w-none">
          <div className="whitespace-pre-wrap leading-relaxed">
            <ReactMarkdown
              remarkPlugins={[remarkGfm]}
              rehypePlugins={[rehypeHighlight]}
              components={{
                code: ({ className, children, ...props }: any) => {
                  const inline = !className?.includes('language-');
                const match = /language-(\w+)/.exec(className || '');
                if (!inline && match) {
                  return (
                    <div className="relative">
                      <pre className="bg-gray-100 rounded-md p-3 overflow-x-auto">
                        <code className={className} {...props}>
                          {children}
                        </code>
                      </pre>
                      <Button
                        size="sm"
                        variant="ghost"
                        className="absolute top-2 right-2 h-6 w-6 p-0"
                        onClick={() => navigator.clipboard.writeText(String(children))}
                      >
                        <Copy className="h-3 w-3" />
                      </Button>
                    </div>
                  );
                }
                return (
                  <code className="bg-gray-100 px-1 py-0.5 rounded text-sm" {...props}>
                    {children}
                  </code>
                );
              },
              pre: ({ children }) => (
                <pre className="bg-gray-100 rounded-md p-3 overflow-x-auto">
                  {children}
                </pre>
              ),
              }}
            >
              {message.content}
            </ReactMarkdown>
          </div>
        </div>

        {/* Render data results and SQL */}
        {message.metadata?.data && (
          <div className="mt-3">
            {renderDataResult(message.metadata.data)}
          </div>
        )}
        {message.metadata?.sql && (
          <div className="mt-3">
            {renderSQL(message.metadata.sql)}
          </div>
        )}
        
        {!isUser && (
          <div className="flex items-center gap-1 mt-3 opacity-0 hover:opacity-100 transition-opacity">
            <Button
              variant="ghost"
              size="sm"
              onClick={handleCopy}
              className="h-7 px-2 text-gray-600 hover:text-gray-800 hover:bg-gray-200"
            >
              <Copy className="h-3 w-3" />
            </Button>
            {onRegenerate && (
              <Button
                variant="ghost"
                size="sm"
                onClick={onRegenerate}
                className="h-7 px-2 text-gray-600 hover:text-gray-800 hover:bg-gray-200"
              >
                <RotateCcw className="h-3 w-3" />
              </Button>
            )}
            {onEdit && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onEdit(message.content)}
                className="h-7 px-2 text-gray-600 hover:text-gray-800 hover:bg-gray-200"
              >
                <Edit className="h-3 w-3" />
              </Button>
            )}
          </div>
        )}
      </div>
      
      {isUser && (
        <div className="flex-shrink-0 w-8 h-8 rounded-full bg-gray-300 flex items-center justify-center text-gray-700 text-sm font-semibold">
          U
        </div>
      )}
    </div>
  );
}
