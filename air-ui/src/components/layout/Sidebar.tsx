import { } from 'react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { 
  BarChart3, 
  MessageSquare, 
  Settings, 
  HelpCircle, 
  Plus,
  ChevronLeft,
  ChevronRight
} from 'lucide-react';
import nubeLogo from '@/assets/nube-logo.png';

interface SidebarProps {
  isCollapsed: boolean;
  onToggle: () => void;
}

export function Sidebar({ isCollapsed, onToggle }: SidebarProps) {
  const menuItems = [
    { icon: MessageSquare, label: 'AI Chat', href: '/chat' },
    { icon: BarChart3, label: 'Reports', href: '/reports' },
    { icon: BarChart3, label: 'Files', href: '/files' },
    { icon: Settings, label: 'Settings', href: '/settings' },
    { icon: HelpCircle, label: 'Help', href: '/help' },
  ];

  return (
    <div className={cn(
      "bg-card border-r transition-all duration-300 flex flex-col",
      isCollapsed ? "w-16" : "w-64"
    )}>
      {/* Header */}
      <div className="p-4 border-b">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <img 
              src={nubeLogo} 
              alt="Nube Logo" 
              className="h-24 w-24 object-contain rounded-lg"
            />
          </div>
          <Button
            variant="ghost"
            size="icon"
            onClick={onToggle}
            className="h-8 w-8"
          >
            {isCollapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}
          </Button>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="p-4 border-b">
        <Button className="w-full justify-start" size="sm">
          <Plus className="h-4 w-4 mr-2" />
          {!isCollapsed && "New Report"}
        </Button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-4">
        <ul className="space-y-2">
          {menuItems.map((item) => (
            <li key={item.label}>
              <a
                variant="ghost"
                className="w-full justify-start inline-flex items-center text-left px-3 py-2 rounded-md hover:bg-accent"
                href={item.href}
              >
                <item.icon className="h-4 w-4 mr-2" />
                {!isCollapsed && item.label}
              </a>
            </li>
          ))}
        </ul>
      </nav>

      {/* Footer */}
      <div className="p-4 border-t">
        <div className="text-xs text-muted-foreground">
          {!isCollapsed && "AIR v1.0.0"}
        </div>
      </div>
    </div>
  );
}
