import { Controller, Control } from 'react-hook-form';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Label } from '@/components/ui/label';
import type { SchemaProperty } from '@/types/api';

interface DynamicFieldProps {
  name: string;
  field: SchemaProperty;
  control: Control<any>;
  error?: string;
}

export function DynamicField({ name, field, control, error }: DynamicFieldProps) {
  const renderField = () => {
    switch (field.type) {
      case 'string':
        if (field.enum && field.enum.length > 0) {
          return (
            <Controller
              name={name}
              control={control}
              render={({ field: fieldProps }) => (
                <Select onValueChange={fieldProps.onChange} value={fieldProps.value}>
                  <SelectTrigger>
                    <SelectValue placeholder={`Select ${field.title || name}`} />
                  </SelectTrigger>
                  <SelectContent>
                    {field.enum?.map((option) => (
                      <SelectItem key={option} value={option}>
                        {option}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            />
          );
        }
        
        if (field.format === 'date') {
          return (
            <Controller
              name={name}
              control={control}
              render={({ field: fieldProps }) => (
                <Input
                  type="date"
                  {...fieldProps}
                  placeholder={field.title || name}
                />
              )}
            />
          );
        }
        
        return (
          <Controller
            name={name}
            control={control}
            render={({ field: fieldProps }) => (
              <Input
                {...fieldProps}
                placeholder={field.title || name}
              />
            )}
          />
        );
        
      case 'number':
      case 'integer':
        return (
          <Controller
            name={name}
            control={control}
            render={({ field: fieldProps }) => (
              <Input
                type="number"
                {...fieldProps}
                placeholder={field.title || name}
                min={field.minimum}
                max={field.maximum}
                onChange={(e) => fieldProps.onChange(parseFloat(e.target.value) || 0)}
              />
            )}
          />
        );
        
      case 'boolean':
        return (
          <Controller
            name={name}
            control={control}
            render={({ field: fieldProps }) => (
              <div className="flex items-center space-x-2">
                <input
                  type="checkbox"
                  checked={fieldProps.value || false}
                  onChange={(e) => fieldProps.onChange(e.target.checked)}
                  className="rounded border-gray-300"
                />
                <Label>{field.title || name}</Label>
              </div>
            )}
          />
        );
        
      default:
        return (
          <Controller
            name={name}
            control={control}
            render={({ field: fieldProps }) => (
              <Input
                {...fieldProps}
                placeholder={field.title || name}
              />
            )}
          />
        );
    }
  };

  return (
    <div className="space-y-2">
      <Label htmlFor={name}>
        {field.title || name}
        {field.description && (
          <span className="text-sm text-muted-foreground ml-2">
            - {field.description}
          </span>
        )}
      </Label>
      {renderField()}
      {error && (
        <p className="text-sm text-destructive">{error}</p>
      )}
    </div>
  );
}
